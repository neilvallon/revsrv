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

type comparableHandler struct {
	http.Handler
}

var undefinedServer = &comparableHandler{
	http.NotFoundHandler(),
}

func main() {
	aliases := getAliases("alias.csv")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()

		host := aliases.Rewrite(r.Host)

		fs := fileServer(host)
		if fs == undefinedServer {
			// wildcard server rewrite
			fs = fileServer(aliases.Rewrite("*"))
		}

		fs.ServeHTTP(w, r)

		// log request
		hostRW := r.Host
		if host != r.Host {
			hostRW = fmt.Sprintf("%s -> %s", r.Host, host)
		}

		log.Printf("[%s] %s in %s", hostRW, r.URL.Path, time.Now().Sub(t1))
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}

func fileServer(host string) *comparableHandler {
	l := strings.Split(host, ".")
	parts := make([]string, len(l))
	for i := range parts {
		parts[i] = l[len(l)-i-1]
	}

	root := path.Join("./", strings.Join(parts, "/"))
	if f, err := os.Stat(root); os.IsNotExist(err) || !f.IsDir() {
		return undefinedServer
	}

	return &comparableHandler{http.FileServer(http.Dir(root))}
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
