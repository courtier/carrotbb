package templates

import (
	"html/template"
	"net/http"

	"github.com/courtier/carrotbb/database"
)

const profilePageTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>carrotbb</title>
</head>

<body>
    {{if .SignedIn}}
    <p>carrotbb - logged in as <a href="/self">{{.Username}}</a> <a href="/createpost">create a post</a> <a href="/logout">log out</a></p>
    {{else}}
    <p>carrotbb - <a href="/signup">sign up</a> <a href="/signin">sign in</a></p>
    {{end}}
	{{if .Posts}}
	<h3>posts</h3>
	<ul>
        {{range .Posts}}
        <li>
			<p><a href="/post/{{.ID}}">{{.Title}}</a> {{ $length := len .CommentIDs }} {{ if ne $length 1 }} {{ $length }} comments {{else}} 1 comment {{end}}, posted at {{.DateCreated.Format "15:04:05 UTC"}} on {{.DateCreated.Format "Jan 02, 2006"}}</p>
        </li>
        {{end}}
    </ul>
	{{else}}
	<h3>no posts found.</h3>
	{{end}}
</body>

</html>`

type ProfilePageTemplateData struct {
	User database.User
}

var (
	profilePageTemplate = template.Must(template.New("profilePageTemplate").Parse(profilePageTemplateStr))
)

func GenerateProfilePage(w http.ResponseWriter, user database.User) error {
	data := ProfilePageTemplateData{
		User: user,
	}
	return profilePageTemplate.Execute(w, data)
}
