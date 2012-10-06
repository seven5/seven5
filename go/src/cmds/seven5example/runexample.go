package main

import (
	"seven5"
	"fmt"
	"net/http"
)

//rest resource for a single city, properties must be public for JSON encoder
type ItalianCity struct {
	Id int32
	Name string
	Population int
	Province string
}

//sample data to work with... so no need for DB
var cityData = []*ItalianCity{
	&ItalianCity{Id:0,Name:"Turin", Province:"Piedmont", Population:900569},
	&ItalianCity{1,"Milan", 3083955, "Lombardy"},
	&ItalianCity{2,"Genoa",800709,"Liguria"},
}


//rest resource for the city list, no data used internally because it is stateless
type ItalianCitiesResource struct{
}
//rest resource for a particular city
type ItalianCityResource struct{
}

//you can use the request parameter to do pagination of a large set of items and such
//but we ignore both headers and query parameters in this func
func (STATELESS *ItalianCitiesResource) Index(IGNORED_headers map[string]string, 
	IGNORED_qp map[string]string) (string,*seven5.Error) {
	return seven5.JsonResult(cityData,true)
}

//used to create dynamic documentation/api
func (STATELESS *ItalianCitiesResource) IndexDoc() []string {
	return []string{""+
	"The resource `/italiancities/` returns a list of known cities in Italy.  Each element of the list is" +
	"a resource of type italiancity that can be fetched individually at `/italiancity/id`.",
	"italiancities ignores the headers supplied in the GET request.",
	"italiancities ignores the query parameters supplied in the URL to GET.",
	"Ignores headers",
	"Ignores query parameters",
	}
}

//given an id, find the object it referencs and return JSON for it.  This ignores
//the headers and query parameters.
func (STATELESS *ItalianCityResource) Find(id int64, hdrs map[string]string, 
	query map[string]string) (string,*seven5.Error) {
	if id<0 || id>=int64(len(cityData)) {
		return seven5.BadRequest(fmt.Sprintf("id must be from 0 to %d",len(cityData)-1))
	}
	return seven5.JsonResult(cityData[id],true)
}

//used to generate documentation/api
func (STATELESS *ItalianCityResource) FindDoc() []string {
	return []string{""+
	"A resource representing a specific italian city at `/italiancity/123`.",
	"Ignores headers",
	"Ignores query parameters",
	}
}

func main() {
	
	h := seven5.NewSimpleHandler()
	h.AddFindAndIndex("italiancity", &ItalianCityResource{},
		"italiancities", &ItalianCitiesResource{}, ItalianCity{})
	
	//this is the _same object_ as h, but just using a different type to make
	//it more "clean" when used with the built in http package.
	asHttp:=seven5.AddDefaultLayout(h)
	
	//normal http calls for running a server in go
	err:=http.ListenAndServe(":3003",logHTTP(asHttp))
	//http.Handle("/dart/", http.StripPrefix("/dart/", http.FileServer(http.Dir("/Users/iansmith/cap/fun/modena/dart"))))
	//fmt.Printf("%+v\n\n", http.DefaultServeMux)
	//err:=http.ListenAndServe(":3003",nil)
	fmt.Printf("Error! Returned from ListenAndServe(): %s", err)
}

// tiny wrapper around all the HTTP dispatching that can be nice to help with debugging
func logHTTP(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}
