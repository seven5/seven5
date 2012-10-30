package italy

import (
	"github.com/seven5/seven5"
	//"seven5"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//sub structure used for a latitude, longitude pair
type LatLng struct {
	Latitude  seven5.Floating
	Longitude seven5.Floating
}

//rest resource for a single city, properties must be public for JSON encoder
type ItalianCity struct {
	Id         seven5.Id
	Name       seven5.String255
	Population seven5.Integer
	Province   seven5.String255
	Location   *LatLng
}

//haveCity returns <0 to mean not found, otherwise returns index of city found.
func haveCity(id seven5.Id) int {
	for i, c := range cityData {
		if c.Id == id {
			return i
		}
	}
	return -828
}

//remove city at an index
func deleteCity(index int, slice []*ItalianCity) []*ItalianCity {
	switch index {
	case 0:
		slice = slice[1:]
	case len(slice) - 1:
		slice = slice[:len(slice)-1]
	default:
		slice = append(slice[:index], slice[index+1:]...)
	}
	return slice
}

//sample data to work with... so no need for DB
var cityData = []*ItalianCity{
	&ItalianCity{Id: 0, Name: "Turin", Province: "Piedmont", Population: 900569,
		Location: &LatLng{45.066667, 7.7}},
	&ItalianCity{1, "Milan", 3083955, "Lombardy", &LatLng{45.464167, 9.190278}},
	&ItalianCity{2, "Genoa", 800709, "Liguria", &LatLng{44.411111, 8.932778}},
}

var cityCount = 3 //we init with 0,1,2 and don't re-use

//rest resource for a particular city, STATELESS!
type ItalianCityResource struct {
}

//Index returns a list of italian cities, filtered by the prefix header and the maximum
//number returned controlled by the max parameter.  
func (STATELESS *ItalianCityResource) Index(headers map[string]string,
	qp map[string]string) (string, *seven5.Error) {

	result := []*ItalianCity{}
	prefix, hasPrefix := headers["Prefix"] //note the capital is always there on headers
	maxStr, hasMax := qp["max"]
	var max int
	var err error

	if hasMax {
		if max, err = strconv.Atoi(maxStr); err != nil {
			return seven5.BadRequest(fmt.Sprintf("can't undestand max parameter %s", maxStr))
		}
	}
	for _, v := range cityData {
		if hasPrefix && !strings.HasPrefix(string(v.Name), prefix) {
			continue
		}
		result = append(result, v)
		if hasMax && len(result) == max {
			break
		}
	}
	return seven5.JsonResult(result, true)
}

//used to create dynamic documentation/api
func (STATELESS *ItalianCityResource) IndexDoc() []string {
	return []string{"" +
		"The resource `/italiancities/` returns a list of known cities in Italy.  Each element of the list is" +
		"a resource of type italiancity that can be fetched individually at `/italiancity/id`.",
		"italiancities ignores the headers supplied in the GET request.",
		"italiancities ignores the query parameters supplied in the URL to GET.",

		"The resource /italiancities/ understands the header 'prefix' and if this header is supplied " +
			"only cities whose Name field begins with the prefix given will be returned.",

		"The resource /italiancities/ allows a query parameter 'max' to control the maximum number " +
			"of cities returned.  No guarantee is made about the order of the returned items. Max must " +
			"be a positive integer (not zero).",
	}
}

//given an id, find the object it referencs and return JSON for it. This ignores
//the query parameters but understands the header 'Round' for rounding pop figures to
//100K boundaries.
func (STATELESS *ItalianCityResource) Find(id seven5.Id, hdrs map[string]string,
	query map[string]string) (string, *seven5.Error) {

	r, hasRound := hdrs["Round"] //note the capital is always there on headers
	n := int64(id)
	if n < 0 || n >= int64(len(cityData)) {
		return seven5.BadRequest(fmt.Sprintf("id must be from 0 to %d", len(cityData)-1))
	}
	pop := cityData[id].Population
	if hasRound && strings.ToLower(r) == "true" {
		excess := cityData[id].Population % 100000
		pop -= excess
		if excess >= 50000 {
			pop += 100000
		}
	}
	data := cityData[id]
	forClient := &ItalianCity{data.Id, data.Name, pop, data.Province, data.Location}
	return seven5.JsonResult(forClient, true)
}

//used to generate documentation/api
func (STATELESS *ItalianCityResource) FindDoc() []string {
	return []string{"" +
		"A resource representing a specific italian city at `/italiancity/123`.",
		"The header 'Round' can be used to get population values rounded to the nearest 100K." +
			"Legal values are true, false, and omitted (which means false).",
		"Ignores query parameters.",
	}
}

//Returns the values at the time of the deletion if successful.
func (STATELESS *ItalianCityResource) Delete(id seven5.Id, headers map[string]string, queryParams map[string]string) (string, *seven5.Error) {
	destroy := haveCity(id)

	if destroy < 0 {
		return seven5.NotFound()
	}

	target := cityData[destroy]
	cityData = deleteCity(destroy, cityData)

	return seven5.JsonResult(&target, true)
}

//Find returns doc for respectively, returned values, accepted headers, and accepted query parameters.
//Three total entries in the resulting slice of strings.  Strings can and should be markdown encoded.
func (STATELESS *ItalianCityResource) DeleteDoc() []string {
	return []string{"We return an instance of ItalianCity if the delete was successful.",
		"Headers are ignored.",
		"Query Parameters are ignored.",
	}
}

func (STATELESS *ItalianCityResource) validateCityData(body string, isPut bool, prev *ItalianCity) (*ItalianCity, string) {
	var bodyCity ItalianCity
	dec := json.NewDecoder(strings.NewReader(body))

	if err := dec.Decode(&bodyCity); err != nil {
		return nil, "Could not understand json body"
	}

	if isPut {
		if bodyCity.Id < 0 {
			return nil, "Must provide an id in body payload of PUT!"
		}
	} else {
		if bodyCity.Id >= 0 {
			return nil, "Cannot supply the Id in POST! Server assigns the Id field!"
		}
	}

	bodyCity.Name = seven5.TrimSpace(bodyCity.Name)
	if bodyCity.Name == "" {
		if isPut {
			bodyCity.Name = prev.Name
		} else {
			return nil, "No name of city provided!"
		}
	}

	bodyCity.Province = seven5.TrimSpace(bodyCity.Province)
	if bodyCity.Province == "" {
		if isPut {
			bodyCity.Province = prev.Province
		} else {
			return nil, "No province of city provided!"
		}
	}

	if bodyCity.Population < 1 {
		if isPut {
			bodyCity.Population = prev.Population
		} else {
			return nil, "No population of city provided!"
		}
	}

	if bodyCity.Location.Latitude < -90.0 || bodyCity.Location.Latitude > 90.0 {
		if isPut {
			bodyCity.Location.Latitude = prev.Location.Latitude
		} else {
			return nil, "Out of bounds latitude!"
		}
	}

	if bodyCity.Location.Longitude < -180.0 || bodyCity.Location.Longitude > 180.0 {
		if isPut {
			bodyCity.Location.Longitude = prev.Location.Longitude
		} else {
			return nil, "Out of bounds longitude!"
		}
	}

	return &bodyCity, ""
}

//Poster takes the object in the body and tries to create a new instance from it.  The resulting instance
//is returned if successful.
func (STATELESS *ItalianCityResource) Post(headers map[string]string, queryParams map[string]string, body string) (string, *seven5.Error) {
	//we re-validate the fields because it's possible the client is doing something nasty and is too lazy
	//or just evil, and is sendnig us bad data
	bodyCity, err := STATELESS.validateCityData(body, false, nil)
	if err != "" {
		return seven5.BadRequest(err)
	}
	//body city is now ready, except needs an id
	bodyCity.Id = seven5.Id(cityCount)
	cityCount++

	cityData = append(cityData, bodyCity)

	return seven5.JsonResult(&bodyCity, true)
}

//Three total entries in the resulting slice of strings.  Strings can and should be markdown encoded.
func (STATELESS *ItalianCityResource) PostDoc() []string {
	return []string{"We return an instance of ItalianCity if the create was successful.",
		"Headers are ignored.",
		"Query Parameters are ignored.",
		"This body should be json for an italian city with all the fields of the resource filled in, " +
			"except the Id which will be assigned in this method.",
	}
}

//Puter takes the object in the body and tries to update the object with the fields provided.
func (STATELESS *ItalianCityResource) Put(id seven5.Id, headers map[string]string, queryParams map[string]string,
	body string) (string, *seven5.Error) {

	var city *ItalianCity
	for _, cand := range cityData {
		if cand.Id == id {
			city = cand
			break
		}
	}
	if city == nil {
		return seven5.NotFound()
	}

	//we re-validate the fields because it's possible the client is doing something nasty and is too lazy
	//or just evil, and is sendnig us bad data
	bodyCity, err := STATELESS.validateCityData(body, true, city)
	if err != "" {
		return seven5.BadRequest(err)
	}
	//body city is now ready, except needs an id
	*city = *bodyCity
	return seven5.JsonResult(&bodyCity, true)
}

//Three total entries in the resulting slice of strings.  Strings can and should be markdown encoded.
func (STATELESS *ItalianCityResource) PutDoc() []string {
	return []string{
		"We the full set of values for the object if the change is successful.",
		"Headers are ignored.",
		"Query Parameters are ignored.",
		"The body must have an Id field that matches the id used in the URL.  Values that are omitted (zero valued) " +
			"are not changed.",
	}
}
