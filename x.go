package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
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

	currReposLog, err := os.Open("repos.txt")
	defer currReposLog.Close()
	switch {
	case err != nil && !errors.Is(err, os.ErrNotExist):
		log.Fatalf("opening repos.txt: %w", err)
	case errors.Is(err, os.ErrNotExist):
		currReposLog, err = os.Create("repos.txt")
		if err != nil {
			log.Fatalf("creating repos.txt: %w", err)
		}
	default:
		s := bufio.NewScanner(currReposLog)
		for s.Scan() {
			_ = os.Remove(s.Text())
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatalf("opening repos.txt: %w", err)
	}
	if !errors.Is(err, os.ErrNotExist) {
	}
	currReposLog.Truncate(0)

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
		os.WriteFile(repo["name"].(string), buff.Bytes(), 0644)
		currReposLog.WriteString(repo["name"].(string) + "\n")
	}
}
