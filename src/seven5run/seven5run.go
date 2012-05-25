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
	log:= &seven5.HtmlLogger{Level: seven5.DEBUG, Writer: writer, Proto: true}
	run(writer,request,log)
}

func run(writer http.ResponseWriter, request *http.Request, log *seven5.HtmlLogger) {
	log.Debug("hello world");
	log.Info("hello world2");
	log.Warn("hello world3");
	log.Error("hello world4");
	log.Dump(request)
}