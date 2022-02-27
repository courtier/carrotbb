package database

import (
	"errors"
	"path/filepath"
	"time"
)

type DatebaseBackends int

const (
	JSON DatebaseBackends = iota
)

type Database interface {
	AddPost(content string, posterID string) error
	AddComment(content string, postID, posterID string) error
	AddUser(name, password, signature string) error

	GetPost(id string) (*Post, error)
	GetComment(id string) (*Comment, error)
	GetUser(id string) (*User, error)
	FindUserByName(name string) (*User, error)

	Disconnect() error
}

type Post struct {
	Content     string
	PosterID    string
	ID          string
	CommentIDs  []string
	DateCreated time.Time
}

type Comment struct {
	Content     string
	PostID      string
	PosterID    string
	ID          string
	DateCreated time.Time
}

type User struct {
	Name       string
	Signature  string
	ID         string
	Password   string
	DateJoined time.Time
}

type DBFrontend struct {
	Backend Database
}

// Connects to the specified backend
// args order is db username, password, address, port, db name
func Connect(backend DatebaseBackends, args ...string) (Database, error) {
	switch backend {
	case JSON:
		folders := filepath.Join("carrotbb", "storage")
		js, err := ConnectJSON(folders, "database.json", 5*time.Minute)
		if err != nil {
			return nil, err
		}
		return &DBFrontend{Backend: js}, nil
	default:
		return nil, errors.New("unsupported database backend")
	}
}

func (db *DBFrontend) AddPost(content string, posterID string) error {
	if err := IsContentValid(content); err != nil {
		return err
	}
	if _, err := db.Backend.GetUser(posterID); err != nil {
		return err
	}
	return db.Backend.AddPost(content, posterID)
}

func (db *DBFrontend) AddComment(content string, postID, posterID string) error {
	if err := IsContentValid(content); err != nil {
		return err
	}
	if _, err := db.Backend.GetPost(postID); err != nil {
		return err
	}
	if _, err := db.Backend.GetUser(posterID); err != nil {
		return err
	}
	return db.Backend.AddComment(content, postID, posterID)
}

func (db *DBFrontend) AddUser(name, password, signature string) error {
	if err := IsUsernameValid(name); err != nil {
		return err
	}
	if _, err := db.FindUserByName(name); err == nil {
		return errors.New("username is taken")
	}
	if err := IsSignatureValid(signature); err != nil {
		return err
	}
	hashedP := HashPassword(name, password)
	return db.Backend.AddUser(name, hashedP, signature)
}

func (db *DBFrontend) GetPost(id string) (*Post, error) {
	return db.Backend.GetPost(id)
}

func (db *DBFrontend) GetComment(id string) (*Comment, error) {
	return db.Backend.GetComment(id)
}

func (db *DBFrontend) GetUser(id string) (*User, error) {
	return db.Backend.GetUser(id)
}

func (db *DBFrontend) FindUserByName(name string) (*User, error) {
	return db.Backend.FindUserByName(name)
}

func (db *DBFrontend) Disconnect() error {
	return db.Backend.Disconnect()
}
