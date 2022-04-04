package templates

import "github.com/courtier/carrotbb/database"

type Profile struct {
	User database.User
	// valid user?
	OK bool
}
