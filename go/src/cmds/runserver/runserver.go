package main

import (
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"modena"
	"os"
	//alias rest2go to rest because that is nicer
	rest "github.com/Kissaki/rest2go"
	"net/http"
)

func main() {
	logger := log.New(os.Stdout, "runserver ", log.LstdFlags)

	address := "127.0.0.1:3003"

	db, err := modena.DBFromEnv(logger)
	if err != nil {
		logger.Fatalf("Cannot find sqlite3 database: %s", err)
	}

	//run statement after defer when stack frame gets popped
	defer db.Close()

	//we init gorp with Sqlite dialect
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	//create the resource
	q := modena.NewQuoteResource(dbmap, logger)
	rest.Resource("/quote/", q)

	//static content :  /static->modena/web
	modenaStaticContent("/static/", "web", logger)
	//static content :  /dart->modena/dart
	modenaStaticContent("/dart/", "dart", logger)

	//function should never return... this is the standard HTTP listener
	if err = http.ListenAndServe(address, logHTTP(http.DefaultServeMux, logger)); err != nil {
		logger.Fatalf("ListenAndServe failed! %s", err)
	}
}

// tiny wrapper around all the HTTP dispatching that can be nice to help with debugging
func logHTTP(handler http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

//serve static content from a given subdir of this modena project
func modenaStaticContent(urlPath string, subdir string, logger *log.Logger) {
	//setup static content
	truePath, err := modena.ModenaPathFromEnv(subdir, logger)
	if err != nil {
		logger.Fatalf("Cannot get path to %s: %s", subdir, err)
	}

	//strip the path from requests so that /urlPath/fart = modena/subdir/fart
	http.Handle(urlPath, http.StripPrefix(urlPath, http.FileServer(http.Dir(truePath))))

}
