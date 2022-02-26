package database

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type JSONDatabaseStructure struct {
	Posts    []Post
	Comments []Comment
	Users    []User
}

type JSONDatabase struct {
	JSONDatabaseStructure
}

func ConnectJSON() (*JSONDatabase, error) {
	jsonFile, err := os.Open(filepath.Join("carrotbb", "storage", "database.json"))
	if err != nil {
		return nil, err
	}
	content, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, err
	}
	dbStructure := JSONDatabaseStructure{}
	err = json.Unmarshal(content, &dbStructure)
	if err != nil {
		return nil, err
	}
	data := &JSONDatabase{JSONDatabaseStructure: dbStructure}
	return data, nil
}

func (j *JSONDatabase) AddPost(content string, categoryID, posterID int) error {
	return nil
}

func (j *JSONDatabase) AddComment(content string, postID, posterID int) error {
	return nil
}

func (j *JSONDatabase) AddUser(name, password, signature string) error {
	return nil
}

func (j *JSONDatabase) GetPost(id int) (Post, error) {
	return Post{}, nil
}

func (j *JSONDatabase) GetComment(id int) (Comment, error) {
	return Comment{}, nil
}

func (j *JSONDatabase) GetUser(id int) (User, error) {
	return User{}, nil
}
