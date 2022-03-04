package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/courtier/carrotbb/database"
	"github.com/courtier/carrotbb/templates"
	"github.com/joho/godotenv"
	"github.com/rs/xid"
)

var (
	db           database.Database
	sessionCache = make(map[string]session)
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		panic(err)
	}

	dbBackend := os.Getenv("DB_BACKEND")

	db, err = database.Connect(dbBackend, 5*time.Minute)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := db.Disconnect(); err != nil {
			log.Println(err)
		}
	}()

	log.Println("connected to database")

	mux := http.NewServeMux()
	mux.HandleFunc("/", IndexPage)

	mux.HandleFunc("/createpost", CreatePostHandler)
	mux.HandleFunc("/post/", PostPage)

	mux.HandleFunc("/createcomment", CreateCommentHandler)

	mux.HandleFunc("/signup", SignupHandler)
	mux.HandleFunc("/signin", SigninHandler)
	mux.HandleFunc("/logout", LogoutHandler)

	auther := NewAuthMiddleware(mux)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Println("listening on port :8080")
		log.Fatal(http.ListenAndServe(":8080", auther))
	}()

	<-terminate
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	posts, err := db.AllPosts()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, err.Error())
		return
	}
	signedIn, username := extractUsername(r)
	if err := templates.GenerateIndexPage(w, signedIn, username, posts); err != nil {
		log.Println(err)
	}
}

func PostPage(w http.ResponseWriter, r *http.Request) {
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
		log.Println(err)
		return
	}
	signedIn, username := extractUsername(r)
	if err := templates.GeneratePostPage(w, signedIn, username, post, poster, comments, users); err != nil {
		log.Println(err)
	}
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		templates.GenerateSignupTemplate(w, r.Referer())
	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "error parsing form")
			log.Println(err)
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
			log.Println(err)
			return
		}
		token, err := newRandomToken()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error generating session token")
			log.Println(err)
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
	if isRequestAuthenticatedSimple(r) {
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
			log.Println(err)
			return
		}
		name := r.Form.Get("username")
		password := r.Form.Get("password")
		redirect := r.Form.Get("redirect")
		hashedP := saltAndHash(password, name)
		user, err := db.FindUserByName(name)
		if err != nil {
			if err == database.ErrUsernameNotFound {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			templates.GenerateErrorPage(w, "error finding that username")
			log.Println(err)
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
			log.Println(err)
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
	if !isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	token, err := extractSession(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error logging out")
		log.Println(err)
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
	if !isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		templates.ServeCreatePostTemplate(w, r)
	case "POST":
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			templates.GenerateErrorPage(w, "error parsing form")
			log.Println(err)
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
		user, err := extractUser(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error extracting user")
			log.Println(err)
			return
		}
		postID, err := db.AddPost(title, content, user.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			templates.GenerateErrorPage(w, "error creating post")
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/post/"+postID.String(), http.StatusFound)
	default:
		w.Header().Add("Allow", "GET, POST")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if !isRequestAuthenticatedSimple(r) {
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
		log.Println(err)
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
		return
	}
	user, err := extractUser(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error extracting user")
		log.Println(err)
		return
	}
	_, err = db.AddComment(content, postID, user.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		templates.GenerateErrorPage(w, "error creating comment")
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/post/"+postID.String(), http.StatusFound)
}

func pathIntoArray(path string) []string {
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return strings.Split(path, "/")
}
