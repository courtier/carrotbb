package database

import (
	"errors"
	"path/filepath"
	"time"

	"github.com/rs/xid"
)

type DatebaseBackends int

const (
	JSON DatebaseBackends = iota
)

type Database interface {
	AddPost(title, content string, posterID xid.ID) (xid.ID, error)
	AddComment(content string, postID, posterID xid.ID) (xid.ID, error)
	AddUser(name, password string) (xid.ID, error)

	GetPost(id xid.ID) (*Post, error)
	GetComment(id xid.ID) (*Comment, error)
	GetUser(id xid.ID) (*User, error)
	FindUserByName(name string) (*User, error)

	AllPosts() ([]Post, error)
	AllCommentsUnderPost(postID xid.ID) ([]Comment, error)

	MapCommentsToUsers(comments []Comment) map[Comment]User

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
	DateCreated time.Time
}

type User struct {
	Name       string
	ID         xid.ID
	Password   string
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

// Connects to the specified backend
// args order is db username, password, address, port, db name
// args order for json is time.Duration for the save interval
func Connect(backend DatebaseBackends, args ...interface{}) (Database, error) {
	var db Database
	switch backend {
	case JSON:
		folders := filepath.Join("carrotbb", "storage")
		interval := 5 * time.Minute
		if len(args) > 0 {
			interval = args[0].(time.Duration)
		}
		js, err := ConnectJSON(folders, "database.json", interval)
		if err != nil {
			return nil, err
		}
		db = js
	default:
		return nil, errors.New("unsupported database backend")
	}
	return db, nil
}
