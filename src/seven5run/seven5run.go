package main

import (
	"net/http"
	"log"
	"time"
	"seven5"
)

func main() {
	s := &http.Server{
		Addr:           ":9009",
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
	}
	http.HandleFunc("/seven5", seven5run)
	log.Fatal(s.ListenAndServe());
}

func seven5run(writer http.ResponseWriter, request *http.Request) {
	bootstrapper := &seven5.Bootstrap{Writer:writer, Request:request}
	bootstrapper.Run();
}
