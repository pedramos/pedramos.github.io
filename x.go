package main

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

const indexTPL = `<html>
    <head>
    	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
        <meta name="go-import" content="plramos.win/{{index . "name"}} git https://github.com/pedramos/{{index . "name"}}">
        <meta http-equiv="refresh" content="0;URL='https://pkg.go.dev/plramos.win/{{index . "name"}}'">
    </head>
    <body>
        Redirecting you to the <a href="https://pkg.go.dev/plramos.win/{{index . "name"}}">go doc page</a>...
    </body>
</html>
`

func main() {
	resp, err := http.Get("https://api.github.com/users/pedramos/repos")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var repos []map[string]any
	err = json.Unmarshal(body, &repos)
	if err != nil {
		log.Fatal(err)
	}

	os.RemoveAll("docs/")
	os.MkdirAll("docs", 0777)

	t := template.Must(template.New("content").Parse(indexTPL))
	for _, repo := range repos {
		if repo["language"] != "Go" {
			continue
		}
		var buff bytes.Buffer
		err := t.Execute(&buff, repo)
		if err != nil {
			log.Fatal(err)
		}
		os.MkdirAll("docs/"+repo["name"].(string), 0777)
		os.WriteFile("docs/"+repo["name"].(string)+"/index.html", buff.Bytes(), 0644)
	}
}
