package database

import (
	"path/filepath"
	"testing"
	"time"
)

func TestConnectJSON(t *testing.T) {
	path := filepath.Join("carrotbb", "storage")
	j, err := ConnectJSON(path, "testdatabase.json", 3*time.Second)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if err = j.AddUser("courtier", "courtier", "hello!"); err != nil {
		t.Log(err)
		t.FailNow()
	}
	time.Sleep(5 * time.Second)
	j.Disconnect()
	j, err = ConnectJSON(path, "testdatabase.json", 30*time.Second)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(j.Users) != 1 {
		t.Log("could not marshal database")
		t.FailNow()
	}
}
