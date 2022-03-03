package database

import (
	"errors"
	"time"

	"github.com/rs/xid"
)

const (
	JSON     string = "json"
	Postgres string = "postgres"
)

var (
	ErrUnsupportedDatabaseBackend = errors.New("unsupported database backend")
	ErrNoPostFoundByID            = errors.New("no matching post id found")
	ErrNoCommentFoundByID         = errors.New("no matching comment id found")
	ErrNoUserFoundByID            = errors.New("no matching user id found")
	ErrUsernameNotFound           = errors.New("no matching user name found")
)

type Database interface {
	AddPost(title, content string, posterID xid.ID) (xid.ID, error)
	AddComment(content string, postID, posterID xid.ID) (xid.ID, error)
	AddUser(name, password string) (xid.ID, error)
	AddSession(tokenHash string, userID xid.ID, expiry time.Time) error

	GetPost(id xid.ID) (Post, error)
	GetComment(id xid.ID) (Comment, error)
	GetUser(id xid.ID) (User, error)
	FindUserByName(name string) (User, error)

	AllPosts() ([]Post, error)

	// TODO: with actual database these all could be 1 query
	GetPostPageData(postID xid.ID) (Post, User, []Comment, map[xid.ID]User, error)

	Disconnect() error
}

type Post struct {
	Title       string
	Content     string
	PosterID    xid.ID
	ID          xid.ID
	CommentIDs  []xid.ID
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

type Session struct {
	TokenHash string
	UserID    xid.ID
	Expiry    time.Time
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

// Connects to the specified backend
// json arg = time.Duration for the save interval
func Connect(backend string, args ...interface{}) (Database, error) {
	switch backend {
	case JSON:
		interval := 5 * time.Minute
		if len(args) > 0 {
			interval = args[0].(time.Duration)
		}
		js, err := ConnectJSON(interval)
		if err != nil {
			return nil, err
		}
		return js, nil
	case Postgres:
		return ConnectPostgres()
	default:
		return nil, ErrUnsupportedDatabaseBackend
	}
}
