package templates

import (
	"fmt"
	"net/http"
)

const createPostTemplate = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CarrotBB Create a Post</title>
</head>

<body>
    <h1>create a post</h1>
    <form action="/createpost" method="post">
        <label for="title">Title</label><br>
        <input type="text" id="title" name="title" placeholder="carrot"/><br>
        <label for="content">Content</label><br>
        <textarea rows="10" cols="80" type="text" id="content" name="content"></textarea><br>
        <input type="submit" value="Submit">
    </form>
</body>

</html>`

func GenerateCreatePostPage() string {
	return createPostTemplate
}

func ServeCreatePostTemplate(w http.ResponseWriter) {
	w.Header().Add("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, createPostTemplate)
}
