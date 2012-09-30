package main

import (
	"net/http"
	"seven5"
	"fmt"
	"log"
)

//rest resource for a single city, properties must be public for JSON encoder
type ItalianCityResource struct {
	Id int32
	Name string
	Population int
	Province string
}

//sample data to work with... so no need for DB
var cityData = []*ItalianCityResource{
	&ItalianCityResource{Id:0,Name:"Turin", Province:"Piedmont", Population:900569},
	&ItalianCityResource{1,"Milan", 3083955, "Lombardy"},
	&ItalianCityResource{2,"Genoa",800709,"Liguria"},
}


//rest resource for the city list, no data use internally because it is stateless
type ItalianCitiesResource struct{
}

//you can use the request parameter to do pagination of a large set of items and such, but we ignore it here
func (STATELESS *ItalianCitiesResource) Index(IGNORED_headers map[string]string, 
	IGNORED_qp map[string]string) (string,*seven5.Error) {
	return seven5.JsonResult(cityData,true)
}

func (STATELESS *ItalianCitiesResource) IndexDoc() (string,string,string) {
	return ""+
"The resource `/italiancities/` returns a list of known cities in Italy.  Each element of the list is" +
"a resource of type italiancity that can be fetched individually at `/italiancity/id`.",
"italiancities ignores the headers supplied in the GET request.",
"italiancities ignores the query parameters supplied in the URL to GET."
}

func (STATELESS *ItalianCityResource) Find(id int32, hdrs map[string]string, 
	query map[string]string) (string,*seven5.Error) {
	if id<0 || id>=int32(len(cityData)) {
		return seven5.BadRequest(fmt.Sprintf("id must be from 0 to %d",len(cityData)-1))
	}
	return seven5.JsonResult(cityData[id],true)
}

func (STATELESS *ItalianCityResource) FindDoc() string {
	return ""+
"A resource representing an italian city."
}

func (STATELESS *ItalianCityResource) FindFields() map[string]*seven5.FieldDoc  {
	return map[string]*seven5.FieldDoc {
		"Name": &seven5.FieldDoc{"","Name of the city, in English"},
		"Population": &seven5.FieldDoc{int(0),"Number of inhabitants, courtesy of Wikipedia"},
		"Province": &seven5.FieldDoc{"","Name of province containing city, in English"},
	}
}

func main() {
	seven5.CurrentHandler.AddResource("italiancities", &ItalianCitiesResource{})
	seven5.CurrentHandler.AddResource("italiancity", &ItalianCityResource{})
	
	err:=http.ListenAndServe(":8080",nil)
	log.Printf("Error! Returned from ListenAndServe(): %s", err)
}