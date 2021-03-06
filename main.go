package main

import (
	"context"
	"crypto/tls"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/courtier/carrotbb/database"
	"github.com/courtier/carrotbb/templates"
	"github.com/joho/godotenv"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

var (
	db     database.Database
	zapper *zap.Logger
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		panic(err)
	}

	zapper, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer zapper.Sync()

	dbBackend := os.Getenv("DB_BACKEND")
	httpPort := os.Getenv("HTTP_PORT")
	httpsPort := os.Getenv("HTTPS_PORT")
	certFile := os.Getenv("SSL_CERT_FILE")
	keyFile := os.Getenv("SSL_KEY_FILE")
	domain := os.Getenv("DOMAIN")

	if httpPort != "" && httpPort[0] != ':' {
		httpPort = ":" + httpPort
	}
	if httpsPort != "" && httpsPort[0] != ':' {
		httpsPort = ":" + httpsPort
	}

	db, err = database.Connect(dbBackend)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Disconnect(); err != nil {
			zapper.Error("error", zap.Error(err))
		}
	}()

	zapper.Info("connected to database", zap.String("backend", dbBackend))

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexPageHandler)
	mux.HandleFunc("/createpost", CreatePostHandler)
	mux.HandleFunc("/post/", PostPageHandler)
	mux.HandleFunc("/createcomment", CreateCommentHandler)
	mux.HandleFunc("/signup", SignupHandler)
	mux.HandleFunc("/signin", SigninHandler)
	mux.HandleFunc("/logout", LogoutHandler)
	mux.HandleFunc("/self", ProfilePageHandler)
	mux.HandleFunc("/user", ProfilePageHandler)

	auther := NewAuthMiddleware(mux)
	logger := NewLoggerMiddleware(auther, zapper)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	if httpPort != "" {
		go func() {
			zapper.Info("listening http", zap.String("port", httpPort))
			zapper.Fatal("http error", zap.Error(http.ListenAndServe(httpPort, logger)))
		}()
	}
	if httpsPort != "" {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			Cache:      autocert.DirCache("/cert-cache"),
			HostPolicy: autocert.HostWhitelist(domain),
		}

		server := &http.Server{
			Addr:    httpsPort,
			Handler: logger,
			TLSConfig: &tls.Config{
				GetCertificate: certManager.GetCertificate,
			},
		}

		go func() {
			zapper.Info("listening https", zap.String("port", httpsPort))
			zapper.Fatal("https error", zap.Error(server.ListenAndServeTLS(certFile, keyFile)))
		}()
	}

	<-terminate
}

func IndexPageHandler(w http.ResponseWriter, r *http.Request) {
	posts, err := db.PagePosts(0, 50)
	if err != nil {
		zapper.Error("error", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, err.Error())
		return
	}
	if err := templates.GenerateIndexPage(w, profileFromCtx(r.Context()), posts); err != nil {
		zapper.Error("error", zap.Error(err))
	}
}

func PostPageHandler(w http.ResponseWriter, r *http.Request) {
	pathSplit := pathIntoArray(r.URL.EscapedPath())
	if len(pathSplit) != 2 || pathSplit[0] != "post" {
		w.WriteHeader(http.StatusBadRequest)
		templates.GenerateErrorPage(w, "malformed request path")
		return
	}
	postID, err := xid.FromString(pathSplit[1])
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.GenerateErrorPage(w, "malformed post id")
		return
	}
	post, poster, comments, users, err := db.GetPostPageData(postID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		templates.GenerateErrorPage(w, "error while fetching the post")
		zapper.Error("error", zap.Error(err))
		return
	}
	if err := templates.GeneratePostPage(w, profileFromCtx(r.Context()), post, poster, comments, users); err != nil {
		zapper.Error("error", zap.Error(err))
	}
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if profileFromCtx(r.Context()).OK {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		if err := templates.GenerateSignupTemplate(w, r.Referer()); err != nil {
			zapper.Error("error", zap.Error(err))
		}
	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "error parsing form")
			zapper.Error("error", zap.Error(err))
			return
		}
		name := r.Form.Get("username")
		password := r.Form.Get("password")
		redirect := r.Form.Get("redirect")
		if err := isUsernameValid(name); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, err.Error())
			return
		}
		if _, err := db.FindUserByName(name); err == nil {
			w.WriteHeader(http.StatusConflict)
			templates.GenerateErrorPage(w, "username is taken")
			return
		}
		if err := isPasswordValid(password); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, err.Error())
			return
		}
		hashedP := saltAndHash(password, name)
		userID, err := db.AddUser(name, hashedP)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error during signup")
			zapper.Error("error", zap.Error(err))
			return
		}
		token, err := newRandomToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error generating session token")
			zapper.Error("error", zap.Error(err))
			return
		}
		authenticateUser(w, token, userID)
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusFound)
	default:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if profileFromCtx(r.Context()).OK {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		templates.GenerateSigninTemplate(w, r.Referer())
	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "error parsing form")
			zapper.Error("error", zap.Error(err))
			return
		}
		name := r.Form.Get("username")
		password := r.Form.Get("password")
		redirect := r.Form.Get("redirect")
		hashedP := saltAndHash(password, name)
		user, err := db.FindUserByName(name)
		if err != nil {
			if err == database.ErrNoUserFoundByName {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			templates.GenerateErrorPage(w, "error finding that username")
			zapper.Error("error", zap.Error(err))
			return
		}
		if hashedP != user.Password {
			w.WriteHeader(http.StatusUnauthorized)
			templates.GenerateErrorPage(w, "incorrect password")
			return
		}
		token, err := newRandomToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error generating session token")
			zapper.Error("error", zap.Error(err))
			return
		}
		authenticateUser(w, token, user.ID)
		if redirect == "" {
			redirect = "/"
		}
		http.Redirect(w, r, redirect, http.StatusFound)
	default:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TODO: if the session cookie is old and is not in cache
// and we log in again, this will return unauthorized
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if !profileFromCtx(r.Context()).OK {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	token, err := extractSessionToken(r)
	if err != nil {
		// This is an internal server error, because we have already
		// checked if the user is authenticated and yet we cannot
		// extract the session token....
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error logging out")
		zapper.Error("error", zap.Error(err))
		return
	}
	unauthenticateUser(w, token)
	var redirect string
	if redirect = r.Referer(); redirect == "" {
		redirect = "/"
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	if !profileFromCtx(r.Context()).OK {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		templates.ServeCreatePostTemplate(w)
	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "error parsing form")
			zapper.Error("error", zap.Error(err))
			return
		}
		title := r.Form.Get("title")
		content := r.Form.Get("content")
		if err := isTitleValid(title); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, err.Error())
			return
		}
		if err := isContentValid(content); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, err.Error())
			return
		}
		// This has to be OK as we  already check for it.
		profile := profileFromCtx(r.Context())
		postID, err := db.AddPost(title, content, profile.User.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error creating post")
			zapper.Error("error", zap.Error(err))
			return
		}
		http.Redirect(w, r, "/post/"+postID.String(), http.StatusFound)
	default:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if !profileFromCtx(r.Context()).OK {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if r.Method != "POST" {
		w.Header().Add("Allow", "POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.GenerateErrorPage(w, "error parsing form")
		zapper.Error("error", zap.Error(err))
		return
	}
	postIDString := r.Form.Get("postID")
	postID, err := xid.FromString(postIDString)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.GenerateErrorPage(w, "malformed post id")
		return
	}
	content := r.Form.Get("comment")
	if err := isContentValid(content); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		templates.GenerateErrorPage(w, err.Error())
		zapper.Error("error", zap.Error(err))
		return
	}
	// Has to be OK.
	profile := profileFromCtx(r.Context())
	_, err = db.AddComment(content, postID, profile.User.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error creating comment")
		zapper.Error("error", zap.Error(err))
		return
	}
	http.Redirect(w, r, "/post/"+postID.String(), http.StatusFound)
}

func ProfilePageHandler(w http.ResponseWriter, r *http.Request) {
	if !profileFromCtx(r.Context()).OK && r.URL.EscapedPath() == "/self" {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	// TODO: add POST for editing profile
	if r.Method != "GET" {
		w.Header().Add("Allow", "GET")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var user templates.Profile
	var err error
	if r.URL.EscapedPath() == "/self" {
		user = profileFromCtx(r.Context())
	} else {
		pathSplit := pathIntoArray(r.URL.EscapedPath())
		if len(pathSplit) != 2 || pathSplit[0] != "user" {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "malformed request path")
			zapper.Error("error", zap.Error(err))
			return
		}
		var userID xid.ID
		userID, err = xid.FromString(pathSplit[1])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "malformed user id")
			zapper.Error("error", zap.Error(err))
			return
		}
		dbUser, err := db.GetUser(userID)
		user = templates.Profile{User: dbUser, OK: err == nil}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error getting user")
		zapper.Error("error", zap.Error(err))
		return
	}
	if err = templates.GenerateProfilePage(w, user); err != nil {
		zapper.Error("error", zap.Error(err))
	}
}

func pathIntoArray(path string) []string {
	if path == "" {
		return []string{}
	}
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return strings.Split(path, "/")
}

// profileFromCtx extracts and returns a templates.Profile struct
// from a context, ideally a request context.
func profileFromCtx(c context.Context) templates.Profile {
	user, ok := c.Value(ContextString("user")).(database.User)
	return templates.Profile{
		User: user,
		OK:   ok,
	}
}
