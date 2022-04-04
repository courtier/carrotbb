package templates

import (
	"html/template"
	"net/http"
)

const profilePageTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>carrotbb</title>
</head>

<body>
    {{if .User.OK}}
    <p>carrotbb - logged in as <a href="/self">{{.User.User.Name}}</a> <a href="/createpost">create a post</a> <a href="/logout">log out</a></p>
	<p>You have created <b>TODO</b> posts.</p>
	<p>You have left <b>TODO</b> comments.</p>
    {{else}}
    <p>carrotbb - <a href="/signup">sign up</a> <a href="/signin">sign in</a></p>
    {{end}}
</body>

</html>`

type ProfilePageTemplateData struct {
	User Profile
}

var (
	profilePageTemplate = template.Must(template.New("profilePageTemplate").Parse(profilePageTemplateStr))
)

// TODO: add links to all created posts, and comments
func GenerateProfilePage(w http.ResponseWriter, user Profile) error {
	data := ProfilePageTemplateData{
		User: user,
	}
	return profilePageTemplate.Execute(w, data)
}
