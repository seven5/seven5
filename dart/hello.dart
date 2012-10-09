#import("dart:json");
#import("dart:html");
#import("seven5.dart");

//generated code derived from the go code
#source("../generated/dart");

main() {
	//note that Index() is a class method (static) because it referencs the entire
	//collection, not an instance
  ItalianCity.Index(dumpAll);
	
	//newly constructed objects are empty, you must use Find() to load their content
	//from the server
  ItalianCity city = new ItalianCity();
	city.Find(0,dumpCity);

	//run simple tests
	exerciseAPI();
}

dumpAll(List<ItalianCity> cities, HttpRequest result) {
	print("number of cities returned from Index: ${cities.length}");
	for (ItalianCity city in cities) {
		print("    city returned from Index(): [${city.Id}] ${city.Name}");
	}
	print("result of 'Index' (GET): ${result.status} ${result.statusText}");
}

dumpCity(ItalianCity cityFound, HttpRequest result) {
	print("city returned from Find() was ${cityFound.Name} with Id ${cityFound.Id}");
	print("object that was found was ${cityFound}");
	print("result of 'Find' (GET) was ${result.status} ${result.statusText}");
}


// print the raw json response text from the server
onSuccess(HttpRequest req) {
	List<Map<String,dynamic>> quotes;
	
	//this is a get of the property 'responseText'
	quotes = JSON.parse(req.responseText); 
	
	Element quoteDiv = new DivElement();
	Element quoteList = new DListElement();
	
	for (Map<String,dynamic> map in quotes) {
		String text = map['Text'];
		String attribution = map['Attribution'];
		
		Element dd = new Element.tag("dd");
		Element dt = new Element.tag("dt");
		
		dt.text = text;
		dd.text = attribution;
		
		quoteList.nodes.add(dt);
		quoteList.nodes.add(dd);
	}
	
	quoteDiv.nodes.add(quoteList);
	
	//just put in body for now
	document.query('body').nodes.add(quoteDiv);
}

/*---------------------------------------------------------------------------*/
/*---------------------------       TEST STUFF      -------------------------*/
/*---------------------------------------------------------------------------*/

exerciseAPI() {
	//exercise the API a bit
	ItalianCity.Index(checkSizeOfCityList(3));
	//on the go side, this will be normalized to "Prefix" with a capital P
	ItalianCity.Index(checkSizeOfCityList(1), null, {"prefix":"T"});

	ItalianCity.Index(checkSizeOfCityList(2), null, null, {"max": "    2"});
	ItalianCity.Index(checkSizeOfCityList(1), null, {"prefix":"T"}, {"max": 2});
	ItalianCity.Index(checkSizeOfCityList(0), null, {"prefix":"Q"}, {"max": "0002"});

	//bogus query parameter
	ItalianCity.Index(null, (response){
		assert(response.status==400);
	}, null, {"max": "two"});


	//
	// TEST FIND
	//
	ItalianCity genoa = new ItalianCity();
	genoa.Find(0, (city, result) {
		assert(city.Name!="Genoa");
	});
	genoa.Find(16, null, (result) {
		assert(result.status==400);
	});
	checkGenoaPopulation(genoa, false, 800709);
	checkGenoaPopulation(genoa, true, 800000);

}

checkGenoaPopulation(ItalianCity genoa, bool rounded, int people) {
	genoa.Find(2, (city, result) {
		assert(city.Name=="Genoa");
		assert(city.Population==people);
	}, null, {"round": "${rounded}"});
}

checkSizeOfCityList(int expectedSize) {
	return (List<ItalianCity> cities, HttpRequest result) {
		print("city length is ${cities.length}");
		assert(cities.length==expectedSize);
		assert(result.status==200);
	};
}
