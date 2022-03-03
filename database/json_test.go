package database

import (
	"os"
	"testing"
	"time"
)

func TestConnectJSON(t *testing.T) {
	os.Setenv("JSON_FOLDER_PATH", "carrotbb/storage")
	os.Setenv("JSON_FILE_NAME", "testdatabase.json")
	j, err := ConnectJSON(3 * time.Second)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if _, err = j.AddUser("courtier", "courtier"); err != nil {
		t.Log(err)
		t.FailNow()
	}
	time.Sleep(5 * time.Second)
	j.Disconnect()
	j, err = ConnectJSON(30 * time.Second)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(j.Users) != 1 {
		t.Log("could not marshal database")
		t.FailNow()
	}
}

func TestSortSliceByDate(t *testing.T) {
	const POST_AMOUNT = 10
	posts := []Post{}
	sortedPosts := make(map[int]Post)
	// fill array with ascending order posts, save the sort in map
	for i := 0; i < POST_AMOUNT; i++ {
		post := Post{DateCreated: time.Now()}
		posts = append(posts, post)
		sortedPosts[i] = post
		time.Sleep(50 * time.Millisecond)
	}
	// reverse array
	for i, j := 0, POST_AMOUNT-1; i < j; i, j = i+1, j-1 {
		temp := posts[i]
		posts[i] = posts[j]
		posts[j] = temp
	}
	// sort
	sortSliceByDate(posts)
	// check
	for i := 0; i < POST_AMOUNT; i++ {
		if posts[i].DateCreated != sortedPosts[i].DateCreated {
			t.FailNow()
		}
	}
}
