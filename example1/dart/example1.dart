#import("dart:json");
#import("dart:html");

//static code that is always needed ... seven5 support library
#import("/seven5/seven5.dart");

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

//getMapsAPIKeey uses the seven5 publicsetting machinery to try to get the api key associated with the
//google maps api.
getMapsAPIKey(List<ItalianCity> cities) {
	HttpRequest req = new HttpRequest();
	req.open("GET", "/seven5/publicsetting/italy/google-maps-api-key");
	req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
		if (req.status != 200) {
			print("request for google maps api key failed... check example1/go/src/italy/publicsetting.json!");
			print("\t server replied with ${req.statusText}");
			buildUI(cities,null);
		} else {
			buildUI(cities, req.responseText);
		}
	});
	req.send();
}

//entry point for building the cities... it's called once we receive the list of cities
//from the server
buildUI(List<ItalianCity> cities, String key) {
	//need a container so twitter can do layout
	DivElement container = new DivElement();
	container.classes.add("container"); //twitter bootstrap
	
	//title spans whole page (width 9 grid units)
	DivElement allCitiesDiv = new DivElement();
	allCitiesDiv.attributes["id"]="all-cities";
	allCitiesDiv.classes.add('row'); //twitter bootstrap
	container.nodes.add(allCitiesDiv);
	
	//span most of the width of the page (spanning element)
	DivElement wideDiv = new DivElement();
	wideDiv.classes.add('span9'); //twitter bootstrap
	allCitiesDiv.nodes.add(wideDiv);
	
	//h3 is not too big 
	HeadingElement h3=new HeadingElement.h3();
	h3.addText('Cities In Italy');
	wideDiv.nodes.add(h3);
	
	//loop over cities returned
	for (ItalianCity city in cities) {
		//cities are a "row" in grid terms
		DivElement row = new DivElement();
		row.classes.add("row"); //twitter boostrap
		
		//left part is for the text
		DivElement left = new DivElement();
		left.classes.add("span3"); //twitter bootstrap
		
		//create a div for this city
		DivElement cityDiv = new DivElement();
		cityDiv.classes.add("city-div");
		//header for this city
		HeadingElement h4=new HeadingElement.h4();
		h4.addText(city.Name);
		left.nodes.add(h4);

		//text about the city is a dl
		DListElement dl = new DListElement();
		dl.classes.add("dl-horizontal"); //twitter bootstrap
		
		//put in textual facts we got from server
		Element dt = new Element.tag("dt");
		dt.addText("Population");
		Element dd = new Element.tag("dd");
		dd.addText("${city.Population}");
		dl.nodes.add(dt);
		dl.nodes.add(dd);
		
		dt=new Element.tag("dt");
		dt.addText("Province");
		dd = new Element.tag("dd");
		dd.addText("${city.Province}");
		dl.nodes.add(dt);
		dl.nodes.add(dd);
		
		//put the list in place
		left.nodes.add(dl);
		
		//map makes up right element
		DivElement right = new DivElement();
		right.classes.add("span6"); //twitter bootstrap
		if (key!=null) {
			DivElement map = new DivElement();
			Element img = new Element.tag("img");

			double lat=city.Location.Latitude;
			double lng=city.Location.Longitude;
		
			img.attributes["src"]=
		 		"http://maps.googleapis.com/maps/api/staticmap?key=${key}&center=${lat},${lng}&zoom=13&size=500x300&sensor=false";
			map.nodes.add(img);
			right.nodes.add(map);
			print("http://maps.googleapis.com/maps/api/staticmap?key=${key}&center=${lat},${lng}&zoom=13&size=500x300&sensor=false");
		} else {
			right.addText("Check your publicsetting.json file to add Google maps api key");
		}

		//put two columns in row
		row.nodes.add(left);
		row.nodes.add(right);
		
		//put row in parent
		wideDiv.nodes.add(row);
	}
	
	//put top level container in the body
	document.query("body").nodes.add(container);
}

//callback from the query that gets all cities (1st parameter is a list)
dumpAll(List<ItalianCity> cities, HttpRequest result) {
	print("number of cities returned from Index: ${cities.length}");
	for (ItalianCity city in cities) {
		print("    city returned from Index(): [${city.Id}] ${city.Name}");
	}
	print("result of 'Index' (GET): ${result.status} ${result.statusText}");

	getMapsAPIKey(cities);
}

//callback from the query that gets a particular city
dumpCity(ItalianCity cityFound, HttpRequest result) {
	print("city returned from Find() was ${cityFound.Name} with Id ${cityFound.Id}");
	print("object that was found was ${cityFound}");
	print("result of 'Find' (GET) was ${result.status} ${result.statusText}");
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
