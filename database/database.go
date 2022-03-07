package database

import (
	"errors"
	"time"

	"github.com/rs/xid"
)

var (
	ErrUnsupportedDatabaseBackend = errors.New("unsupported database backend")
	ErrNoPostFoundByID            = errors.New("no matching post id found")
	ErrNoCommentFoundByID         = errors.New("no matching comment id found")
	ErrNoUserFoundByID            = errors.New("no matching user id found")
	ErrNoUserFoundByName          = errors.New("no matching user name found")
)

type Database interface {
	// AddPost adds a post to the database
	AddPost(title, content string, posterID xid.ID) (xid.ID, error)
	// AddComment adds a comment to the database
	AddComment(content string, postID, posterID xid.ID) (xid.ID, error)
	// AddUser adds a user to the database
	AddUser(name, password string) (xid.ID, error)

	// GetPost gets a post from the database
	GetPost(id xid.ID) (Post, error)
	// GetComment gets a comment from the database
	GetComment(id xid.ID) (Comment, error)
	// GetUser gets a user from the database
	GetUser(id xid.ID) (User, error)
	// FindUserByName finds a user by that name in the database
	FindUserByName(name string) (User, error)

	// AllPosts returns all the posts in the database
	AllPosts() ([]Post, error)

	// GetPostPageData returns all the data necessary to render a post page
	GetPostPageData(postID xid.ID) (Post, User, []Comment, map[xid.ID]User, error)

	// Disconnect gracefully disconnects from a database
	Disconnect() error
}

type Post struct {
	Title       string
	Content     string
	PosterID    xid.ID
	ID          xid.ID
	CommentIDs  [][]byte
	DateCreated time.Time
}

type Comment struct {
	Content     string
	PostID      xid.ID
	PosterID    xid.ID
	ID          xid.ID
	Deleted     bool
	DateCreated time.Time
}

type User struct {
	Name       string
	ID         xid.ID
	Password   string
	Deleted    bool
	DateJoined time.Time
}

type DBFrontend struct {
	Backend Database
}

var (
	DeletedUser = User{
		Name:       "Deleted",
		ID:         xid.NilID(),
		Password:   "",
		DateJoined: time.Time{},
	}
)

// Connect connects to the specified database backend
// Possible values are "json" and "postgres"
func Connect(backend string) (Database, error) {
	switch backend {
	case "json":
		interval := 5 * time.Minute
		js, err := ConnectJSON(interval)
		if err != nil {
			return nil, err
		}
		return js, nil
	case "postgres":
		return ConnectPostgres()
	default:
		return nil, ErrUnsupportedDatabaseBackend
	}
}
