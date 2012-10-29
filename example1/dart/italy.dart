#import("dart:json");
#import("dart:html");
#import("dart:isolate");

//static code that is always needed ... seven5 support library
#import("/seven5/seven5.dart");

//generated code derived from the go code
#source("../generated/dart");

int mapCounter = 1;
int mapRequestDelay = 250; //millis

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
void getMapsAPIKey(List<ItalianCity> cities) {
	HttpRequest req = new HttpRequest();
	req.open("GET", "/seven5/publicsetting/italy/google-maps-api-key");
	req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
		if (req.status != 200) {
			print("request for google maps api key failed... check example1/go/src/italy/publicsetting.json!");
			print("\t server replied with ${req.statusText}");
			displayUI(cities, null);
		} else {
			print("Got the google maps API key... ${req.responseText}");
			displayUI(cities, req.responseText);
		}
	});
	req.send();
}

//build the parts of the UI that don't change--this only needs to be done once
//this calls the code to display the content of cities and possibly key 
void displayUI(List<ItalianCity> cities, String key) {
	DivElement container = document.query("#outercontainer");
	container.classes.add("container"); //twitter bootstrap
	
	//full width for the dynamic part
	DivElement cityDiv = document.query("#citydiv");
	cityDiv.classes.add('row');

	//title it
	cityDiv.nodes.add(makeTitling('Cities In Italy'));
	
	//container for our display of city rows
	DivElement fullWidthDiv = new DivElement();
	fullWidthDiv.classes.add('span9');
	
	cityDiv.nodes.add(fullWidthDiv);
	
	//static input part of ui
	DivElement inputDiv = document.query("#inputdiv");
	inputDiv.classes.add('row');

	//make a form for the user to enter new cities
	Element form = new Element.tag("form");
	form.classes.add("form-horizontal"); //twitter bootstrap
	form.classes.add("span3"); //twitter bootstrap
	form.classes.add("offset1"); //twitter bootstrap
	
	Element legend = new Element.tag('legend');
	legend.addText('Add Another City');
	form.nodes.add(legend);
	
	//make some input rows
	form.nodes.add(makeInputRow('City Name'));
	form.nodes.add(makeInputRow('Population'));
	form.nodes.add(makeInputRow('Province'));
	form.nodes.add(makeInputRow('Latitude'));
	form.nodes.add(makeInputRow('Longitude'));
	
	DivElement buttonDiv = new DivElement();
	buttonDiv.classes.add("controls");//twitter bootstrap
	
	Element button = new Element.tag("button");
	button.classes.add('btn');//twitter bootstrap
	button.addText("Add");
	
	buttonDiv.nodes.add(button);
	form.nodes.add(buttonDiv);
	
	inputDiv.nodes.add(form);
	
	//ok, we are now ready to dislpay the data we loaded from the server
	displayKnownCities(cities, fullWidthDiv, key);
}

HeadingElement makeTitling(String text) {
	HeadingElement h3=new HeadingElement.h3();
	h3.addText(text);
	h3.classes.add('span12'); //twitter bootstrap
	return h3;
}

DivElement makeInputRow(String label) {
	
	DivElement row = new DivElement();
	row.classes.add("control-group");//twitter bootstrap
	
	String id=label.toLowerCase().replaceAll(" ","");
	
	Element labelElement = new Element.tag('label');
	labelElement.addText(label);
	labelElement.classes.add("control-label");//twitter bootstrap
	labelElement.attributes['for']=id;
	
	DivElement control = new DivElement();
	control.classes.add('controls');
	
	Element input = new Element.tag('input');
	input.attributes['type']='text';
	input.attributes['id']=id;
	input.attributes['placeholder']=label;
	
	control.nodes.add(input);
	
	row.nodes.add(labelElement);
	row.nodes.add(control);
	return row;
}

//entry point for building the cities... it's called once we receive the list of cities
//from the server
void displayKnownCities(List<ItalianCity> cities, DivElement parent, String key) {
	
	//this can be called multiple times so we need to clean out anything laying around
	if (!parent.nodes.isEmpty()) {
		parent.clear();
	}
	
	bool isGray = false;
	
	//loop over cities returned
	for (ItalianCity city in cities) {
		//cities are a "row" in grid terms
		DivElement row = new DivElement();
		row.classes.add("row"); //twitter boostrap
		
		if (isGray) {
			row.classes.add("gray"); //italy.css
			isGray=false;
		} else {
			isGray=true;
		}
		
		//left part is for the text
		DivElement left = new DivElement();
		left.classes.add("span4"); //twitter bootstrap
		
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
		
		//put a delete button there
		Element button = new Element.tag("button");
		button.addText("Delete");
		button.classes.add("btn"); //twitter bootsrap
		button.classes.add("btn-danger"); //twitter bootsrap
		button.classes.add("offset1");
		left.nodes.add(button);
		
		//map makes up right element
		DivElement right = new DivElement();
		right.classes.add("span5"); //twitter bootstrap

		//put two columns in row
		row.nodes.add(left);
		row.nodes.add(right);
		
		//put row in parent
		parent.nodes.add(row);
		
		//if we have a key, try for some maps
		if (key!=null) {
			loadGoogleMap(key, city.Location, right);
		}
	}
}

//you must have a google api key and that key must be turned on for the "staticmaps" (not "maps"!) api
void loadGoogleMap(String key, LatLng loc, DivElement parent) {
	new Timer(mapRequestDelay*mapCounter, (Timer ignored) {
		DivElement map = new DivElement();
		map.classes.add("mapdiv");
		
		double lat=loc.Latitude;
		double lng=loc.Longitude;
		
		int padding = 4; //defined in CSS file, really
		String size = "${380-(2*padding)}x${226-(2*padding)}";
		
		Element img = new Element.tag("img");
		img.attributes["src"]=
	 		"http://maps.googleapis.com/maps/api/staticmap?key=${key}&zoom=10&center=${lat},${lng}&zoom=13&size=${size}&sensor=false";

		map.nodes.add(img);
		parent.nodes.add(map);
	});
	mapCounter++;
}

//callback from the query that gets all cities (1st parameter is a list)
void dumpAll(List<ItalianCity> cities, HttpRequest result) {
	print("number of cities returned from Index: ${cities.length}");
	for (ItalianCity city in cities) {
		print("    city returned from Index(): [${city.Id}] ${city.Name}");
	}
	print("result of 'Index' (GET): ${result.status} ${result.statusText}");

	getMapsAPIKey(cities);
}

//callback from the query that gets a particular city
void dumpCity(ItalianCity cityFound, HttpRequest result) {
	print("city returned from Find() was ${cityFound.Name} with Id ${cityFound.Id}");
	print("object that was found was ${cityFound}");
	print("result of 'Find' (GET) was ${result.status} ${result.statusText}");
}

/*---------------------------------------------------------------------------*/
/*---------------------------       TEST STUFF      -------------------------*/
/*---------------------------------------------------------------------------*/

void exerciseAPI() {
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

void checkGenoaPopulation(ItalianCity genoa, bool rounded, int people) {
	genoa.Find(2, (city, result) {
		assert(city.Name=="Genoa");
		assert(city.Population==people);
	}, null, {"round": "${rounded}"});
}

Function checkSizeOfCityList(int expectedSize) {
	return (List<ItalianCity> cities, HttpRequest result) {
		print("city length is ${cities.length}");
		assert(cities.length==expectedSize);
		assert(result.status==200);
	};
}
