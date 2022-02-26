package database

import (
	"errors"
	"time"
)

type DatebaseBackends int

const (
	JSON DatebaseBackends = iota
)

type Database interface {
	AddPost(content string, categoryID, posterID int) error
	AddComment(content string, postID, posterID int) error
	AddUser(name, password, signature string) error

	GetPost(id int) (Post, error)
	GetComment(id int) (Comment, error)
	GetUser(id int) (User, error)
	FindUserByName(name string) (User, error)
}

type Post struct {
	Content     string
	PosterID    int
	ID          int
	CommentIDs  []int
	DateCreated time.Time
}

type Comment struct {
	Content     string
	PostID      int
	PosterID    int
	ID          int
	DateCreated time.Time
}

type User struct {
	Name       string
	Signature  string
	ID         int
	DateJoined time.Time
}

type DBFrontend struct {
	Backend Database
}

func Connect(backend DatebaseBackends) (Database, error) {
	switch backend {
	case JSON:
		js, err := ConnectJSON()
		if err != nil {
			return nil, err
		}
		return &DBFrontend{Backend: js}, nil
	default:
		return nil, errors.New("unsupported database backend")
	}
}

func (db *DBFrontend) AddPost(content string, categoryID, posterID int) error {
	return nil
}

func (db *DBFrontend) AddComment(content string, postID, posterID int) error {
	return nil
}

func (db *DBFrontend) AddUser(name, password, signature string) error {
	if err := IsUsernameValid(name); err != nil {
		return err
	}
	if _, err := db.FindUserByName(name); err == nil {
		return errors.New("username is taken")
	}
	hashedP := HashPassword(name, password)
	return db.Backend.AddUser(name, hashedP, signature)
}

func (db *DBFrontend) GetPost(id int) (Post, error) {
	return Post{}, nil
}

func (db *DBFrontend) GetComment(id int) (Comment, error) {
	return Comment{}, nil
}

func (db *DBFrontend) GetUser(id int) (User, error) {
	return User{}, nil
}

func (db *DBFrontend) FindUserByName(name string) (User, error) {
	return User{}, nil
}
