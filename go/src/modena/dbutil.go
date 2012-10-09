package modena

import (
	"database/sql"
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//defines a new error type
var BAD_GOPATH = errors.New("GOPATH is not defined or is empty")

//derive DB path from the GOPATH var
func DBFromEnv(logger *log.Logger) (*sql.DB, error) {
	dbpath, err := DBPathFromEnv("modena.sqlite", logger)
	if err != nil {
		return nil, err
	}
	logger.Printf("modena db path: %s", dbpath)

	db, err := sql.Open("sqlite3", dbpath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

//compute a filename in the db directory
func DBPathFromEnv(name string, logger *log.Logger) (string, error) {
	dbdir, err := ModenaPathFromEnv("db", logger)
	if err != nil {
		return "", err
	}
	//using Join to avoid issues with separators... Dir is a textual way to find the parent dir
	return filepath.Join(dbdir, name), nil
}

//compute a directory inside the project level of modena via GOPATH
func ModenaPathFromEnv(rootDir string, logger *log.Logger) (string, error) {
	env := os.Getenv("GOPATH")
	if env == "" {
		return "", BAD_GOPATH
	}
	pieces := strings.Split(env, ":")
	if len(pieces) > 1 {
		logger.Printf("Using first part of GOPATH [%s] to find DB directory", pieces[0])
		env = pieces[0]
	}
	return filepath.Join(filepath.Dir(env), rootDir),nil
}
