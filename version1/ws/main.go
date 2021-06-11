package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[SERVE] %+v", r.URL)
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	PREFIX := "/gopath/"
	found := false
	http.HandleFunc(PREFIX, func(w http.ResponseWriter, r *http.Request) {
		path := os.Getenv("GOPATH")
		pieces := strings.Split(path, ":")
		for _, p := range pieces {
			candidate := filepath.Join(p, r.URL.Path[len(PREFIX):])
			_, err := os.Open(candidate)
			log.Printf("[GOPATH] %s (%v)\n", candidate, err)
			if err == nil {
				//no error, so lets read it
				http.ServeFile(w, r, candidate)
				found = true
				break
			}
		}
		if !found {
			w.WriteHeader(http.StatusNotFound)
		}
	})
	port := 8898
	if os.Getenv("PORT") != "" {
		portraw := os.Getenv("PORT")
		p, err := strconv.ParseInt(portraw, 10, 64)
		if err != nil {
			log.Fatalf("Can't understand PORT environment var: %s", portraw)
		}
		port = int(p)
	}
	log.Printf(fmt.Sprintf("waiting on :%d", port))
	log.Fatalf("%s", http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}
