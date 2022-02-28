package templates

import (
	"html/template"
	"net/http"

	"github.com/courtier/carrotbb/database"
)

const indexPageTemplateStr = `<html lang="en">

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
			<p><a href="/{{.ID}}">{{.Title}}</a> {{len .CommentIDs}} comments, posted at {{.DateCreated.Format "15:04:05 UTC"}} on {{.DateCreated.Format "Jan 02, 2006"}}</p>
        </li>
        {{end}}
    </ul>
	{{else}}
	<h3>no posts found.</h3>
	{{end}}
</body>

</html>`

type IndexPageTemplateData struct {
	SignedIn bool
	Username string
	Posts    []database.Post
}

var (
	indexPageTemplate = template.Must(template.New("indexPageTemplate").Parse(indexPageTemplateStr))
)

func GenerateIndexPage(w http.ResponseWriter, signedIn bool, name string, posts []database.Post) error {
	data := IndexPageTemplateData{
		SignedIn: signedIn,
		Username: name,
		Posts:    posts,
	}
	return indexPageTemplate.Execute(w, data)
}
