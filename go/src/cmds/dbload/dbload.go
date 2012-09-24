package main

import (
	"encoding/json"
	"github.com/coopernurse/gorp"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"modena"
	"os"
)

//note that main takes no args.  there are some routines in "os" package if you want to mess with
//the process level arguments
func main() {
	logger := log.New(os.Stdout, "dbload ", log.LstdFlags)

	db, err := modena.DBFromEnv(logger)
	if err != nil {
		logger.Fatalf("Cannot find sqlite3 database: %s", err)
	}

	//run statement after defer when stack frame gets popped
	defer db.Close()

	//we init gorp with Sqlite dialect
	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.SqliteDialect{}}

	//tell it to work on Quote objects and we intentionally ignore return value
	_ = dbmap.AddTable(modena.Quote{}).SetKeys(true, "Id")

	//destroy all tables!
	if err = dbmap.DropTables(); err != nil {
		logger.Printf("Apparently no tables to drop...")
	} else {
		logger.Printf("Destroying all tables...")
	}

	//create new tables
	if err = dbmap.CreateTables(); err != nil {
		logger.Fatalf("Cannot init tables: %s", err)
	}

	//find fixed data
	quotePath, err := modena.DBPathFromEnv("quote.json", logger)
	if err != nil {
		logger.Fatalf("Unable to compute path to 'quote.json' in 'db' dir:%s", err)
	}

	//get an open File*
	file, err := os.Open(quotePath)
	if err != nil {
		logger.Fatalf("Unable to open 'quote.json': %s",err)
	}
	//run the code after defer when this stack frame is popped
	defer file.Close()
	
	//decode the json data into a slice of Quote objects
	decoder:=json.NewDecoder(file)
	fixedQuote := []modena.Quote{}
	if err=decoder.Decode(&fixedQuote); err!=nil {
		logger.Fatalf("Unable to decode data in quotes.json: %s", err)
	}
	
	logger.Printf("Found %d quotes...loading into tables.", len(fixedQuote))
	for i,q := range fixedQuote {
		if err=dbmap.Insert(&q); err!=nil {
			logger.Fatal("Unable to insert quote #%d into database:",i,err)
		}
	}
	
}
