package templates

import (
	"html/template"
	"net/http"
	"strings"
)

const errorPageTemplateStr = `<html lang="en">

<head>
    <meta charset="UTF-8">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>carrotbb - error</title>
</head>

<body>
    <p><a href="/">carrotbb</a> - you've just run into an error!</p>
	<h4>{{.Error}}</h4>
</body>

</html>`

type ErrorPageTemplateData struct {
	Error string
}

var (
	errorPageTemplate = template.Must(template.New("errorPageTemplate").Parse(errorPageTemplateStr))
)

func GenerateErrorPage(w http.ResponseWriter, args ...string) error {
	data := ErrorPageTemplateData{
		Error: intoOneString(args),
	}
	return errorPageTemplate.Execute(w, data)
}

func intoOneString(args []string) string {
	var sb strings.Builder
	for i, s := range args {
		sb.WriteString(s)
		if i != len(args)-1 {
			sb.WriteByte(' ')
		}
	}
	return sb.String()
}
