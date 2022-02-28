package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/courtier/carrotbb/database"
	"github.com/courtier/carrotbb/templates"
	"github.com/rs/xid"
)

var (
	db           database.Database
	sessionCache = make(map[string]session)
)

func main() {
	var err error
	db, err = database.Connect(database.JSON, 5*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	defer func() {
		if err := db.Disconnect(); err != nil {
			log.Println(err)
		}
	}()

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
		fmt.Fprintln(w, "Trouble with the database:", err)
		return
	}
	var signedIn bool
	var username string
	user, err := extractUser(r)
	if err == nil {
		signedIn = true
		username = user.Name
	}
	templates.GenerateIndexPage(w, signedIn, username, posts)
}

func PostPage(w http.ResponseWriter, r *http.Request) {
	pathSplit := pathIntoArray(r.URL.EscapedPath())
	if pathSplit[len(pathSplit)-2] != "post" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "malformed request path")
		return
	}
	postIDString := pathSplit[len(pathSplit)-1]
	postID, err := xid.FromString(postIDString)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// TODO: with actual database these all could be 1 query
	post, err := db.GetPost(postID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	comments, err := db.AllCommentsUnderPost(postID)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	poster, err := db.GetUser(post.PosterID)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	var signedIn bool
	var username string
	user, err := extractUser(r)
	if err == nil {
		signedIn = true
		username = user.Name
	}
	users := db.MapCommentsToUsers(comments)
	templates.GeneratePostPage(w, signedIn, username, *post, *poster, comments, users)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	switch r.Method {
	case "GET":
		templates.ServeSignupTemplate(w, r)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		name := r.Form.Get("username")
		password := r.Form.Get("password")
		if err := isUsernameValid(name); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if _, err := db.FindUserByName(name); err == nil {
			fmt.Fprintln(w, errors.New("username is taken"))
			return
		}
		if err := isPasswordValid(password); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		hashedP := hashPassword(name, password)
		userID, err := db.AddUser(name, hashedP)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		token, err := newRandomToken()
		if err != nil {
			fmt.Fprintln(w, "error generating session token", err)
			return
		}
		authenticateUser(w, token, userID)
		http.Redirect(w, r, "/", http.StatusFound)
	default:
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
		templates.ServeSigninTemplate(w, r)
	case "POST":
		if err := r.ParseForm(); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		name := r.Form.Get("username")
		password := r.Form.Get("password")
		hashedP := hashPassword(name, password)
		user, err := db.FindUserByName(name)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if user.Password != hashedP {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintln(w, "incorrect password")
			return
		}
		token, err := newRandomToken()
		if err != nil {
			fmt.Fprintln(w, "error generating session token", err)
			return
		}
		authenticateUser(w, token, user.ID)
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	if !isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	token, err := extractSession(r)
	if err != nil {
		fmt.Println(w, err)
		return
	}
	unauthenticateUser(w, token)
	http.Redirect(w, r, "/", http.StatusFound)
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
			fmt.Fprintln(w, err)
			return
		}
		title := r.Form.Get("title")
		content := r.Form.Get("content")
		if err := isTitleValid(title); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		if err := isContentValid(content); err != nil {
			fmt.Fprintln(w, err)
			return
		}
		user, err := extractUser(r)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		postID, err := db.AddPost(title, content, user.ID)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		http.Redirect(w, r, "/post/"+postID.String(), http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TODO:
func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
	if !isRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		fmt.Fprintln(w, err)
		return
	}
	title := r.Form.Get("title")
	content := r.Form.Get("content")
	if err := isTitleValid(title); err != nil {
		fmt.Fprintln(w, err)
		return
	}
	if err := isContentValid(content); err != nil {
		fmt.Fprintln(w, err)
		return
	}
	user, err := extractUser(r)
	if err != nil {
		fmt.Fprintln(w, err)
		return
	}
	postID, err := db.AddPost(title, content, user.ID)
	if err != nil {
		fmt.Fprintln(w, err)
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
