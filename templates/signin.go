package templates

import (
	"html/template"
	"net/http"
)

const signinTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CarrotBB Signin</title>
</head>

<body>
    <h1>sign in to carrotbb</h1>
    <form action="/signin" method="post">
        <label for="username">Username</label><br>
        <input type="text" id="username" name="username" placeholder="carrot"><br>
        <label for="password">Password</label><br>
        <input type="password" id="password" name="password"><br><br>
        <input type="hidden" id="redirect" name="redirect" value="{{ .Redirect }}">
        <input type="submit" value="Submit">
    </form>
</body>

</html>`

type SigninTemplateData struct {
	Redirect string
}

var (
	signinTemplate = template.Must(template.New("signinTemplate").Parse(signinTemplateStr))
)

func GenerateSigninTemplate(w http.ResponseWriter, referer string) error {
	if referer == "" {
		referer = "/"
	}
	data := SigninTemplateData{
		Redirect: referer,
	}
	return signinTemplate.Execute(w, data)
}
