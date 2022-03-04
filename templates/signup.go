package templates

import (
	"html/template"
	"net/http"
)

const signupTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CarrotBB Signup</title>
</head>

<body>
    <h1>sign up to carrotbb</h1>
    <form action="/signup" method="post">
        <label for="username">Username</label><br>
        <input type="text" id="username" name="username" placeholder="carrot"><br>
        <label for="password">Password</label><br>
        <input type="password" id="password" name="password"><br><br>
        <input type="hidden" id="redirect" name="redirect" value="{{ .Redirect }}">
        <input type="submit" value="Submit">
    </form>
</body>

</html>`

type SignupTemplateData struct {
	Redirect string
}

var (
	signupTemplate = template.Must(template.New("signupTemplate").Parse(signupTemplateStr))
)

func GenerateSignupTemplate(w http.ResponseWriter, referer string) error {
	if referer == "" {
		referer = "/"
	}
	data := SignupTemplateData{
		Redirect: referer,
	}
	return signupTemplate.Execute(w, data)
}
