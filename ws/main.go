package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	log.Printf("waiting on :8898")
	log.Fatalf("%s", http.ListenAndServe(":8898", nil))
}
