package italy

//This file is to show you some of the key ideas about testing seven5 apps on the server side.

import (
	"encoding/json"
	"github.com/seven5/seven5"
	"net/http"
	//"seven5"
	"strings"
	"testing"
)

//used when we don't want to supply a query param or header set
var emptyMap = make(map[string]string)

//check result
func checkResultCode(T *testing.T, msg string, result string, err *seven5.Error, expectedStatus int) {
	if expectedStatus/100 == 2 {
		if err != nil {
			if err.StatusCode != expectedStatus {
				T.Errorf("Expected a success (%v) but got (%v) for '%s'", expectedStatus, err.StatusCode, msg)
			}
		}
	} else {
		if err.StatusCode != expectedStatus {
			T.Errorf("Expected status of %v but got %v for '%s'", expectedStatus, err.StatusCode, msg)
		}
	}
	if expectedStatus/100 != 2 {
		if result != "" {
			T.Errorf("Expected an error, but also got a result (%s) for '%s'", result, msg)
		}
	}
}

//NOTE: you can create these server resources anytime/anywhere, because they are stateless
var cityResource = &ItalianCityResource{}

//test that if you give a correct id you don't get an error, but you do get error on bad id
func TestIdNum(T *testing.T) {

	//NOTE: if you pass params here you are _not_ using the web stack but testing the same code path
	//NOTE: (in your code) that seven5 will use when a real web request is processed
	result, err := cityResource.Find(0, emptyMap, emptyMap)
	checkResultCode(T, "id 0 is acceptable", result, err, http.StatusOK)

	result, err = cityResource.Find(214, emptyMap, emptyMap)
	checkResultCode(T, "id 214 is bogus", result, err, http.StatusBadRequest)

}

//utility routine to get the city or list of cities from json
func decodeJson(T *testing.T, jsonBlob string, ptr interface{}) interface{} {
	dec := json.NewDecoder(strings.NewReader(jsonBlob))
	if err := dec.Decode(ptr); err != nil {
		T.Fatalf("Error doing json decoding: %s", err)
	}
	return ptr
}

//make a map for use as a header or query parameter set
func makeMap(key string, value string) map[string]string {
	result := make(map[string]string)
	result[key] = value
	return result
}

//check that a given header with a given value produces a rounded or unrounded result
func checkRoundingVariant(T *testing.T, msg string, key string, value string, rounded bool, id seven5.Id) {
	result, err := cityResource.Find(id, makeMap(key, value), emptyMap)
	if err != nil {
		T.Fatalf("Can't find italian city %v: %s", id, err)
	}
	var city ItalianCity
	city = *(decodeJson(T, result, &city).(*ItalianCity))
	if rounded {
		if city.Population%100000 != 0 {
			T.Errorf("Expected rounded result for id %d but got population of %d for '%s'!", id,
				city.Population, msg)
		}
	} else {
		if city.Population%100000 == 0 {
			T.Errorf("Expected non-rounded result for id %d but got population of %d for '%s'!", id,
				city.Population, msg)
		}
	}
}

//TestRounding checks that the city resource correctly responds to headers passed in the GET
func TestRounding(T *testing.T) {
	//validate that no params works and that the city is suitable for a test of rounding
	id := seven5.Id(0)
	result, err := cityResource.Find(id, emptyMap, emptyMap)
	if err != nil {
		T.Fatalf("Can't find italian city %v: %s", id, err)
	}
	var notRounded ItalianCity
	notRounded = *(decodeJson(T, result, &notRounded).(*ItalianCity))
	if notRounded.Population%100000 == 0 {
		T.Fatalf("Can't use id %v for rounding, no rounding necessary: %v", id, notRounded.Population)
	}
	if notRounded.Population%100000 != 0 {
		//note that go normalizes headers to always start with upper case, so no need to bother with
		//lowercase tests
		checkRoundingVariant(T, "lowercase value", "Round", "true", true, id)
		checkRoundingVariant(T, "bogus key", "Fart", "True", false, id)
		checkRoundingVariant(T, "uppercase value", "Round", "True", true, id)
	}
}

//check that a given query parameter correctly affects the resulting number of results
func checkMaxVariant(T *testing.T, msg string, key string, value string, expected int) {
	result, err := cityResource.Index(emptyMap, makeMap(key, value))
	if err != nil {
		T.Fatalf("Can't index all italian cities for '%s': %s", msg, err)
	}
	var cities []ItalianCity
	cities = *(decodeJson(T, result, &cities).(*[]ItalianCity))
	if len(cities) != expected {
		T.Errorf("Expected %d but got %d results for '%s'!", expected, len(cities), msg)
	}
}

//check that a given header with a given value produces a filtered list of cities
func checkFilteringVariant(T *testing.T, msg string, key string, value string, expected int) {
	result, err := cityResource.Index(makeMap(key, value), emptyMap)
	if err != nil {
		T.Fatalf("Can't index all italian cities for '%s': %s", msg, err)
	}
	var cities []ItalianCity
	cities = *(decodeJson(T, result, &cities).(*[]ItalianCity))
	if len(cities) != expected {
		T.Errorf("Expected %d but got %d results for '%s'!", expected, len(cities), msg)
	}
}

//TestMaxQueryParameter tests that max can limit the number of results and that we correctly
//understand the case-sensitivity of query parameters
func TestMaxQueryParameter(T *testing.T) {
	checkMaxVariant(T, "bogus parameter", "Bogus", "1", 3)
	checkMaxVariant(T, "max of 1, lowercase", "max", "1", 1)
	checkMaxVariant(T, "max of 1, uppercase", "Max", "1", 3)
	checkMaxVariant(T, "max of 2, lowercase", "max", "2", 2)
}

//TestFiltering
func TestFiltering(T *testing.T) {
	checkFilteringVariant(T, "bogus header (Uber)", "Uberbogus", "T", 3)
	checkFilteringVariant(T, "bogus header (Filter)", "Filter", "T", 3)
	checkFilteringVariant(T, "Simple filter (T)", "Prefix", "T", 1)
	checkFilteringVariant(T, "Filters everything", "Prefix", "Z", 0)
}
