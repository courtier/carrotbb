package database

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
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

	postsLock    sync.RWMutex
	commentsLock sync.RWMutex
	usersLock    sync.RWMutex

	saveTicker *time.Ticker
	stopSaving chan bool

	backingPath string
}

func ConnectJSON(saveInterval time.Duration) (*JSONDatabase, error) {
	if os.Getenv("JSON_FOLDER_PATH") == "" || os.Getenv("JSON_FILE_NAME") == "" {
		log.Fatalln("json folder path or file name is empty")
	}
	folders := filepath.FromSlash(os.Getenv("JSON_FOLDER_PATH"))
	fileName := os.Getenv("JSON_FILE_NAME")
	if err := os.MkdirAll(folders, 0777); err != nil {
		return nil, err
	}
	path := filepath.Join(folders, fileName)
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
		saveTicker:            time.NewTicker(saveInterval),
		stopSaving:            make(chan bool, 1),
		backingPath:           path,
	}
	go func() {
		for {
			select {
			case <-data.stopSaving:
				return
			case <-data.saveTicker.C:
				log.Println("saving json database")
				if err := data.saveDatabase(); err != nil {
					log.Println("error saving json database", err)
				}
			}
		}
	}()
	return data, nil
}

func (j *JSONDatabase) saveDatabase() error {
	jsonFile, err := os.OpenFile(j.backingPath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	j.postsLock.Lock()
	j.commentsLock.Lock()
	j.usersLock.Lock()
	defer j.postsLock.Unlock()
	defer j.commentsLock.Unlock()
	defer j.usersLock.Unlock()
	bs, err := json.Marshal(j.JSONDatabaseStructure)
	if err != nil {
		return err
	}
	_, err = jsonFile.Write(bs)
	return err
}

func (j *JSONDatabase) Disconnect() error {
	// TODO: there has to be a race condition here fosho
	// j.postsLock.Lock()
	// j.commentsLock.Lock()
	// j.usersLock.Lock()
	// defer j.postsLock.Unlock()
	// defer j.commentsLock.Unlock()
	// defer j.usersLock.Unlock()
	j.stopSaving <- true
	return j.saveDatabase()
}

func (j *JSONDatabase) AddPost(title, content string, posterID xid.ID) (xid.ID, error) {
	j.postsLock.Lock()
	defer j.postsLock.Unlock()
	newID := xid.New()
	newP := Post{
		Title:       title,
		Content:     content,
		ID:          newID,
		PosterID:    posterID,
		DateCreated: time.Now(),
		CommentIDs:  [][]byte{},
	}
	j.Posts = append(j.Posts, newP)
	return newID, nil
}

func (j *JSONDatabase) AddComment(content string, postID, posterID xid.ID) (xid.ID, error) {
	j.commentsLock.Lock()
	defer j.commentsLock.Unlock()
	newID := xid.New()
	newC := Comment{
		Content:     content,
		ID:          newID,
		PosterID:    posterID,
		PostID:      postID,
		DateCreated: time.Now(),
	}
	j.Comments = append(j.Comments, newC)
	// TODO: this is probably a race condition
	post, err := j.GetPost(postID)
	if err != nil {
		return newID, err
	}
	post.CommentIDs = append(post.CommentIDs, newID.Bytes())
	return newID, nil
}

func (j *JSONDatabase) AddUser(name, password string) (xid.ID, error) {
	j.usersLock.Lock()
	defer j.usersLock.Unlock()
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

func (j *JSONDatabase) GetPost(id xid.ID) (Post, error) {
	j.postsLock.RLock()
	defer j.postsLock.RUnlock()
	for n := range j.Posts {
		if j.Posts[n].ID == id {
			return j.Posts[n], nil
		}
	}
	return Post{}, ErrNoPostFoundByID
}

func (j *JSONDatabase) GetComment(id xid.ID) (Comment, error) {
	j.commentsLock.RLock()
	defer j.commentsLock.RUnlock()
	for n := range j.Comments {
		if j.Comments[n].ID == id {
			return j.Comments[n], nil
		}
	}
	return Comment{}, ErrNoCommentFoundByID
}

func (j *JSONDatabase) GetUser(id xid.ID) (User, error) {
	j.usersLock.RLock()
	defer j.usersLock.RUnlock()
	for n := range j.Users {
		if j.Users[n].ID == id {
			return j.Users[n], nil
		}
	}
	return User{}, ErrNoUserFoundByID
}

func (j *JSONDatabase) FindUserByName(name string) (User, error) {
	j.usersLock.RLock()
	defer j.usersLock.RUnlock()
	for n := range j.Users {
		if j.Users[n].Name == name {
			return j.Users[n], nil
		}
	}
	return User{}, ErrNoUserFoundByName
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

func (j *JSONDatabase) GetPostPageData(postID xid.ID) (post Post, poster User, comments []Comment, users map[xid.ID]User, err error) {
	post, err = j.GetPost(postID)
	if err != nil {
		return
	}
	poster, err = j.GetUser(post.PosterID)
	if err != nil {
		return
	}
	comments = make([]Comment, 0)
	users = make(map[xid.ID]User)
	for _, cID := range post.CommentIDs {
		commentP, err := j.GetComment(xid.Must(xid.FromBytes(cID)))
		if err != nil {
			// TODO: Ignore error here or return out of the function?
			continue
		}
		comments = append(comments, commentP)
	}
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].DateCreated.Before(comments[j].DateCreated)
	})
	for _, comment := range comments {
		commenterP, err := j.GetUser(comment.PosterID)
		if err != nil {
			users[comment.ID] = DeletedUser
			continue
		}
		users[comment.ID] = commenterP
	}
	return
}

func sortSliceByDate(slice interface{}) {
	switch v := slice.(type) {
	case []Post:
		sort.Slice(v, func(i, j int) bool {
			return v[i].DateCreated.Before(v[j].DateCreated)
		})
	case []Comment:
		sort.Slice(v, func(i, j int) bool {
			return v[i].DateCreated.Before(v[j].DateCreated)
		})
	case []User:
		sort.Slice(v, func(i, j int) bool {
			return v[i].DateJoined.Before(v[j].DateJoined)
		})
	}
}
