package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
)

const indexTPL = `<html>
    <head>
    	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
        <meta name="go-import" content="plramos.win/{{.}} git https://github.com/pedramos/{{.}}">
        <meta http-equiv="refresh" content="0;URL='https://pkg.go.dev/plramos.win/{{.}}'">
    </head>
    <body>
        Redirecting you to the <a href="https://pkg.go.dev/plramos.win/{{.}}">go doc page</a>...
    </body>
</html>
`

var (
	LsFlag        = flag.Bool("l", false, "Fetch github list")
	reposFilePath = "repos.csv"
)

func ListsReposGithub() ([]string, error) {
	resp, err := http.Get("https://api.github.com/users/pedramos/repos")
	if err != nil {
		return nil, fmt.Errorf("GET https://api.github.com: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response from https://api.github.com: %w", err)
	}

	var repos []map[string]any
	err = json.Unmarshal(body, &repos)
	if err != nil {
		return nil, fmt.Errorf("parsing response from https://api.github.com: %w", err)
	}
	var targets []string
	for _, repo := range repos {
		if repo["language"] != "Go" {
			continue
		}

		targets = append(targets, repo["name"].(string))

	}
	return targets, nil
}

func ListReposFile() ([]string, error) {
	f, err := os.Open(reposFilePath)
	if err != nil {
		return nil, fmt.Errorf("reading repos file: &w", err)
	}
	s := bufio.NewScanner(f)

	var repos []string
	for s.Scan() {
		repos = append(repos, s.Text())
	}
	return repos, nil
}

func main() {

	flag.Parse()
	if *LsFlag {
		repos, err := ListsReposGithub()
		if err != nil {
			log.Fatal(err)
		}
		for _, r := range repos {
			fmt.Println(r)
		}
	}
	os.RemoveAll("docs/")
	os.MkdirAll("docs", 0777)

	t := template.Must(template.New("content").Parse(indexTPL))
	repos, err := ListReposFile()
	if err != nil {
		log.Fatal(err)
	}
	for _, repo := range repos {
		var buff bytes.Buffer
		err := t.Execute(&buff, repo)
		if err != nil {
			log.Fatalf("building redirect for %s: %w", repo, err)
		}
		os.MkdirAll("docs/"+repo, 0777)
		os.WriteFile("docs/"+repo+"/index.html", buff.Bytes(), 0644)
	}
}
