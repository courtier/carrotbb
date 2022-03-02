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
