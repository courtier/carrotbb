package templates

import (
	"html/template"
	"net/http"

	"github.com/courtier/carrotbb/database"
	"github.com/rs/xid"
)

const postPageTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>carrotbb</title>
</head>

<body>
    {{if .SignedIn}}
    <p><a href="/">carrotbb</a> - logged in as <a href="/self">{{.Username}}</a> <a href="/createpost">create a post</a> <a href="/logout">log out</a></p>
    {{else}}
    <p><a href="/">carrotbb</a> - <a href="/signup">sign up</a> <a href="/signin">sign in</a></p>
    {{end}}
    <p><b>{{.Poster.Name}}</b> posted at {{.Post.DateCreated.Format "15:04:05 UTC"}} on {{.Post.DateCreated.Format "Jan 02, 2006"}}:</p>
	<h2>{{.Post.Title}}</h2>
    <p>{{.Post.Content}}</p>
    <hr>
    {{if .Comments}}
        {{range $comment, $user := .Comments}}
			<p><b>{{$user.Name}}</b> commented at {{$comment.DateCreated.Format "15:04:05 UTC"}} on {{$comment.DateCreated.Format "Jan 02, 2006"}}<br>
                {{$comment.Content}}</p>
            <hr>
        {{end}}
	{{else}}
	<p><b>no comments found.{{if .SignedIn}} leave one down below!{{end}}</b></p>
	{{end}}
    {{if .SignedIn}}
    <form action="/createcomment" method="post">
        <label for="comment">Comment</label><br>
		<input type="hidden" id="postID" name="postID" value="{{.Post.ID}}">
        <textarea rows="7" cols="50" id="comment" name="comment"></textarea><br><br>
        <input type="submit" value="Submit">
    </form>
    {{end}}
</body>

</html>`

type PostPageTemplateData struct {
	SignedIn bool
	Username string
	Post     database.Post
	Poster   database.User
	Comments []database.Comment
	Users    map[xid.ID]database.User
}

var (
	postPageTemplate = template.Must(template.New("postPageTemplate").Parse(postPageTemplateStr))
)

func GeneratePostPage(w http.ResponseWriter, signedIn bool, name string, post database.Post, poster database.User, comments []database.Comment, users map[xid.ID]database.User) error {
	data := PostPageTemplateData{
		SignedIn: signedIn,
		Username: name,
		Post:     post,
		Poster:   poster,
		Comments: comments,
	}
	return postPageTemplate.Execute(w, data)
}
