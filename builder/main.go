package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"

	"rsc.io/gitfs"
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
	lsFlag    = flag.Bool("l", false, "Fetch github list")
	reposFlag = flag.String("r", "repos.csv", "List of repos to include")
	dirFlag   = flag.String("d", "", "Destination for build result")
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
	f, err := os.Open(*reposFlag)
	if err != nil {
		return nil, fmt.Errorf("reading repos file: %w", err)
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
	repos := []string{}

	if *lsFlag {
		var err error
		repos, err = ListsReposGithub()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("found on github")
		for _, r := range repos {
			fmt.Println("	- " + r)
		}
	}

	r, err := ListReposFile()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Fatal(err)
	}
	repos = append(repos, r...)

	if *dirFlag == "" {
		return
	}
	_, err = os.Stat(*dirFlag)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		log.Fatalf("directory %s exists", *dirFlag)
	}

	err = os.MkdirAll(*dirFlag, 0o777)
	if err != nil {
		log.Fatalf("failed to create %s\n\t%s", *dirFlag, err)
	}

	for _, repo := range repos {
		r, err := subPackages(repo)
		if err != nil {
			fmt.Printf("WARN: %s", err)
			continue
		}
		repos = append(repos, r...)
	}

	fmt.Println(repos)

	t := template.Must(template.New("content").Parse(indexTPL))
	for _, repo := range repos {

		var buff bytes.Buffer
		err := t.Execute(&buff, repo)
		if err != nil {
			log.Fatalf("building redirect for %s: %s", repo, err)
		}

		os.MkdirAll(path.Join(*dirFlag, repo), 0o777)
		if err != nil {
			log.Fatalf("failed to create %s\n\t%s", path.Join(*dirFlag, repo), err)
		}
		os.WriteFile(path.Join(*dirFlag, repo, "index.html"), buff.Bytes(), 0o644)
		if err != nil {
			log.Fatalf("failed to create %s\n\t%s", path.Join(*dirFlag, repo, "index.html"), err)
		}
	}
}

func subPackages(repo string) (repos []string, err error) {
	r, err := gitfs.NewRepo("https://github.com/pedramos/" + repo)
	if err != nil {
		return repos, err
	}
	_, repofs, err := r.Clone("HEAD")
	if err != nil {
		return repos, err
	}
	err = fs.WalkDir(repofs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != "." {
			repos = append(repos, repo+"/"+path)
		}
		return nil
	})
	return repos, err
}
