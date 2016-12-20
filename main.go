package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	aliases := getAliases("alias.csv")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		host := aliases.Rewrite(r.Host)

		defer func() {
			// log request
			hostRW := r.Host
			if host != r.Host {
				hostRW = fmt.Sprintf("%s -> %s", r.Host, host)
			}

			log.Printf("[%s] %s in %s", hostRW, r.URL.Path, time.Now().Sub(t1))
		}()

		dir := getRootDir(host)
		if dir == "" {
			// wildcard server rewrite
			defaultDomain := getRootDir(aliases.Rewrite("*"))
			if defaultDomain == "" {
				http.Error(w, "Undefined Server", 444)
			} else {
				http.FileServer(defaultDomain).ServeHTTP(w, r)
			}
			return
		}

		// custom error rewrites
		filePath := path.Join(string(dir), path.Clean("/"+r.URL.Path))
		stat, err := os.Stat(filePath)

		if os.IsNotExist(err) {
			aliases.Error(w, r, dir, http.StatusNotFound)
			return
		}

		if stat.IsDir() {
			if _, err := os.Stat(path.Join(filePath, "/index.html")); os.IsNotExist(err) {
				aliases.Error(w, r, dir, http.StatusForbidden)
				return
			}
		}

		http.FileServer(dir).ServeHTTP(w, r)
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func getRootDir(host string) http.Dir {
	l := strings.Split(host, ".")
	parts := make([]string, len(l))
	for i := range parts {
		parts[i] = l[len(l)-i-1]
	}

	root := path.Join("./", path.Clean("/"+strings.Join(parts, "/")))
	if f, err := os.Stat(root); os.IsNotExist(err) || !f.IsDir() {
		return ""
	}

	dir := http.Dir(root)
	return dir
}

type aliasSet map[string]string

func getAliases(name string) aliasSet {
	f, err := os.Open(name)
	if err != nil {
		return make(aliasSet)
	}
	defer f.Close()

	records, err := csv.NewReader(f).ReadAll()
	if err != nil {
		log.Println(err)
		return make(aliasSet)
	}

	var set = make(aliasSet)
	for _, r := range records {
		from, to := strings.TrimSpace(r[0]), strings.TrimSpace(r[1])
		log.Println("adding alias", from, "->", to)
		set[from] = to
	}

	return set
}

func (a aliasSet) Rewrite(host string) string {
	if to, ok := a[host]; ok {
		host = to
	}
	return host
}

func (a aliasSet) Error(w http.ResponseWriter, r *http.Request, dir http.Dir, code int) {
	var codeText string
	switch code {
	case http.StatusNotFound:
		codeText = "[404]"
	case http.StatusForbidden:
		codeText = "[403]"
	default:
		codeText = "[000]"
	}

	if fileName := a.Rewrite(codeText); fileName == codeText {
		http.Error(w, http.StatusText(code), code)
		return
	} else {
		// got custom error
		r.URL.Path = fileName

		f, err := dir.Open(fileName)
		if err != nil {
			http.Error(w, http.StatusText(code), code)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			http.Error(w, http.StatusText(code), code)
			return
		}

		http.ServeContent(w, r, fileName, stat.ModTime(), f)
	}
}
