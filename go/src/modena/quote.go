package modena

import (
	//base types for dealing with requests and creating responses
	"net/http"
	//alias rest2go to rest because that is nicer
	_ "github.com/Kissaki/rest2go"
	//this is the mapper for pushing/pulling objects to the DB
	"github.com/coopernurse/gorp"
	//basic logging is enough for me
	"bytes"
	"encoding/json"
	"log"
)

// note that the fields have to be uppercase (public) to be seen by reflection API and this
// is needed by the code that emits the sql to save/load these objects
type Quote struct {
	Id          int64 //managed by the DB and GoRP
	Text        string
	Attribution string
}

//object that understands what it means to be this type of resource... database and logger
//are private but there isn't really a good reason for that
type QuoteResource struct {
	dbmap  *gorp.DbMap
	logger *log.Logger
}

// Create a pointer to this type of resource
func New(dbmap *gorp.DbMap, logger *log.Logger) *QuoteResource {
	return &QuoteResource{dbmap, logger}
}

//Indexer is type defined by rest2go... we are implementing that interface here
func (self *QuoteResource) Index(resp http.ResponseWriter) {
	query := "select Id, Text, Attribution from Quote"
	all, err := self.dbmap.Select(Quote{}, query)
	if err != nil {
		http.Error(resp, "500 Internal Error [SQL Failure]", http.StatusInternalServerError)
		self.logger.Printf("Problem reading the list of quotes: %s", err)
		return
	}

	//create a json encoder connected to an auto-growing buffer
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	
	//Encode the objects in the slice into a json string
	if err=encoder.Encode(&all); err!=nil {
		http.Error(resp, "500 Internal Error [JSON encoding failure]", http.StatusInternalServerError)
		self.logger.Printf("Problem encoding json: %s", err)
		return
	}
	
	//push the bytes out over the wire
	if n, err := resp.Write(buffer.Bytes()); err!=nil {
		self.logger.Printf("Problem writing the output bytes (%d): %s",n,err)
	}
	
	
}
