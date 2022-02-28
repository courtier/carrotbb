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
)

var (
	db           database.Database
	sessionCache = make(map[string]Session)
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
	mux.HandleFunc("/post/", PostPage)
	mux.HandleFunc("/", IndexPage)

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
	if len(posts) == 0 {
		fmt.Fprintln(w, "No posts found")
		return
	}
	for _, p := range posts {
		fmt.Fprintln(w, p.Title)
	}
}

func PostPage(w http.ResponseWriter, r *http.Request) {
	pathSplit := pathIntoArray(r.URL.EscapedPath())
	if pathSplit[len(pathSplit)-2] != "post" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "malformed request path")
		return
	}
	postID := pathSplit[len(pathSplit)-1]
	log.Println("received post request with id", postID)
}

func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if IsRequestAuthenticatedSimple(r) {
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
		hashedP := HashPassword(name, password)
		userID, err := db.AddUser(name, hashedP)
		if err != nil {
			fmt.Fprintln(w, err)
			return
		}
		token, err := NewRandomToken()
		if err != nil {
			fmt.Fprintln(w, "error generating session token", err)
			return
		}
		authenticateUser(w, token, userID)
		http.Redirect(w, r, "/", 200)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func SigninHandler(w http.ResponseWriter, r *http.Request) {
	if IsRequestAuthenticatedSimple(r) {
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
		hashedP := HashPassword(name, password)
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
		token, err := NewRandomToken()
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
	if !IsRequestAuthenticatedSimple(r) {
		http.Redirect(w, r, "/", http.StatusUnauthorized)
		return
	}
	token, err := ExtractSession(r)
	if err != nil {
		fmt.Println(w, err)
		return
	}
	delete(sessionCache, token)
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
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
