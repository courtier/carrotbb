package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/courtier/carrotbb/database"
)

var db database.Database

func main() {
	var err error
	db, err = database.Connect(database.JSON, 5*time.Minute)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Disconnect()

	http.HandleFunc("/post/", PostPage)

	http.HandleFunc("/signin", IndexPage)
	http.HandleFunc("/signup", IndexPage)

	http.HandleFunc("/", IndexPage)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		log.Println("listening on port :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	<-terminate
}

func IndexPage(w http.ResponseWriter, r *http.Request) {
	log.Println("received index request")
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

func pathIntoArray(path string) []string {
	if path[0] == '/' {
		path = path[1:]
	}
	if path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return strings.Split(path, "/")
}
