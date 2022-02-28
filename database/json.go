package database

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/xid"
)

type JSONDatabaseStructure struct {
	Posts    []Post
	Comments []Comment
	Users    []User
}

type JSONDatabase struct {
	JSONDatabaseStructure

	PostsLock    sync.RWMutex
	CommentsLock sync.RWMutex
	UsersLock    sync.RWMutex

	SaveTicker *time.Ticker
	StopSaving chan bool

	BackingPath string
}

func ConnectJSON(folders, filename string, saveInterval time.Duration) (*JSONDatabase, error) {
	if err := os.MkdirAll(folders, 0777); err != nil {
		return nil, err
	}
	path := filepath.Join(folders, filename)
	jsonFile, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	defer jsonFile.Close()
	content, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	dbStructure := JSONDatabaseStructure{}
	if len(content) > 0 {
		err = json.Unmarshal(content, &dbStructure)
		if err != nil {
			return nil, err
		}
	}
	data := &JSONDatabase{
		JSONDatabaseStructure: dbStructure,
		SaveTicker:            time.NewTicker(saveInterval),
		StopSaving:            make(chan bool, 1),
		BackingPath:           path,
	}
	go func() {
		for {
			select {
			case <-data.StopSaving:
				return
			case <-data.SaveTicker.C:
				log.Println("saving json database")
				if err := data.SaveDatabase(); err != nil {
					log.Println("error saving json database", err)
				}
			}
		}
	}()
	return data, nil
}

func (j *JSONDatabase) SaveDatabase() error {
	jsonFile, err := os.OpenFile(j.BackingPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	j.PostsLock.Lock()
	j.CommentsLock.Lock()
	j.UsersLock.Lock()
	defer j.PostsLock.Unlock()
	defer j.CommentsLock.Unlock()
	defer j.UsersLock.Unlock()
	bs, err := json.Marshal(j.JSONDatabaseStructure)
	if err != nil {
		return err
	}
	_, err = jsonFile.Write(bs)
	return err
}

func (j *JSONDatabase) Disconnect() error {
	// TODO: there has to be a race condition here fosho
	// j.PostsLock.Lock()
	// j.CommentsLock.Lock()
	// j.UsersLock.Lock()
	// defer j.PostsLock.Unlock()
	// defer j.CommentsLock.Unlock()
	// defer j.UsersLock.Unlock()
	j.StopSaving <- true
	return j.SaveDatabase()
}

func (j *JSONDatabase) AddPost(title, content string, posterID xid.ID) (xid.ID, error) {
	j.PostsLock.Lock()
	defer j.PostsLock.Unlock()
	newID := xid.New()
	newP := Post{
		Title:       title,
		Content:     content,
		ID:          newID,
		PosterID:    posterID,
		DateCreated: time.Now(),
	}
	j.Posts = append(j.Posts, newP)
	return newID, nil
}

func (j *JSONDatabase) AddComment(content string, postID, posterID xid.ID) (xid.ID, error) {
	j.CommentsLock.Lock()
	defer j.CommentsLock.Unlock()
	j.PostsLock.Lock()
	defer j.PostsLock.Unlock()
	newID := xid.New()
	newC := Comment{
		Content:     content,
		ID:          newID,
		PosterID:    posterID,
		PostID:      postID,
		DateCreated: time.Now(),
	}
	j.Comments = append(j.Comments, newC)
	post, err := j.GetPost(postID)
	if err != nil {
		return newID, nil
	}
	post.CommentIDs = append(post.CommentIDs, newID)
	return newID, nil
}

func (j *JSONDatabase) AddUser(name, password string) (xid.ID, error) {
	j.UsersLock.Lock()
	defer j.UsersLock.Unlock()
	newID := xid.New()
	newU := User{
		Name:       name,
		ID:         newID,
		Password:   password,
		DateJoined: time.Now(),
	}
	j.Users = append(j.Users, newU)
	return newID, nil
}

func (j *JSONDatabase) GetPost(id xid.ID) (*Post, error) {
	j.PostsLock.RLock()
	defer j.PostsLock.RUnlock()
	for _, p := range j.Posts {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("no matching post id found")
}

func (j *JSONDatabase) GetComment(id xid.ID) (*Comment, error) {
	j.CommentsLock.RLock()
	defer j.CommentsLock.RUnlock()
	for _, c := range j.Comments {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, errors.New("no matching comment id found")
}

func (j *JSONDatabase) GetUser(id xid.ID) (*User, error) {
	j.UsersLock.RLock()
	defer j.UsersLock.RUnlock()
	for _, u := range j.Users {
		if u.ID == id {
			return &u, nil
		}
	}
	return nil, errors.New("no matching user id found")
}

func (j *JSONDatabase) FindUserByName(name string) (*User, error) {
	j.UsersLock.RLock()
	defer j.UsersLock.RUnlock()
	for _, u := range j.Users {
		if u.Name == name {
			return &u, nil
		}
	}
	return nil, errors.New("no matching user name found")
}

func (j *JSONDatabase) AllPosts() ([]Post, error) {
	return j.Posts, nil
}

func (j *JSONDatabase) AllCommentsUnderPost(postID xid.ID) ([]Comment, error) {
	cs := []Comment{}
	for _, c := range j.Comments {
		if c.PostID == postID {
			cs = append(cs, c)
		}
	}
	return cs, nil
}
