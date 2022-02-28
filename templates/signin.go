package templates

import (
	"fmt"
	"net/http"
)

const signinTemplate = `<html lang="en">

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
        <input type="submit" value="Submit">
    </form>
</body>

</html>`

func GenerateSigninPage() string {
	return signupTemplate
}

func ServeSigninTemplate(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, signinTemplate)
}
