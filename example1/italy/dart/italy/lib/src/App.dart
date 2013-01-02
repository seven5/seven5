library italy;

import 'dart:json';
import 'dart:html';
import 'dart:isolate';
import 'dart:math';
//generated code derived from the go code
import '/generated/dart';

int mapCounter = 1;
int mapRequestDelay = 50; //millis

String mapsAPIKey = null;

main() {
	
	//////////////////////////// SELF TESTS /////////////////////////////////////////
	//note that Index() is a class method (static) because it referencs the entire
	//collection, not an instance
  //ItalianCity.Index(dumpAll);
	
	//newly constructed objects are empty, you must use Find() to load their content
	//from the server
  //ItalianCity city = new ItalianCity();
	//city.Find(0,dumpCity);

	//run simple tests
	//exerciseAPI();
	
	//////////////////////////// REAL APPLICATION UI//////////////////////////////////
	
	//check to see if we have a maps key
	getMapsAPIKey((String key) {
		ItalianCity.Index((List<ItalianCity> cities, HttpRequest result) {
			if (result.status!=200) {
				print("Failed to get italian city list: ${result.statusText}");
			} else {
				displayUI(cities, key);
				mapsAPIKey = key;
			}
		});
	});
}

//getMapsAPIKeey uses the seven5 publicsetting machinery to try to get the api key associated with the
//google maps api.  It is assigned, if returned, to mapsAPIKey.
void getMapsAPIKey(Function fn) {
	HttpRequest req = new HttpRequest();
	req.open("GET", "/seven5/publicsetting/italy/google-maps-api-key");
	req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
		if (req.status != 200) {
			print("request for google maps api key failed... check example1/go/src/italy/publicsetting.json!");
			print("\t server replied with ${req.statusText}");
			fn(null);
		} else {
			print("Got the google maps API key... ${req.responseText}");
			fn(req.responseText);
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
	form.attributes['onSubmit']="return false;"; //we're doing the love in DART
	
	Element legend = new Element.tag('legend');
	legend.text='Add Another City';
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
	button.text="Add";

	button.on.click.add((Event event) {		
		ItalianCity c = new ItalianCity();
		c.Location = new LatLng();
		
		//create an italian city example from the form fields
		c.Name = document.query("#cityname").value;
		c.Province = document.query("#province").value;
		c.Population = int.parse(document.query("#population").value);
		c.Location.Latitude = double.parse(document.query("#latitude").value);
		c.Location.Longitude = double.parse(document.query("#longitude").value);
		c.Id = -19008; //just a signal
		
		//send the example to the server (no id)
		ItalianCity.Post(c, (newObject, response) {
			print("created new city with instance id ${newObject.Id}");
			//create was successful, so redraw list of cities
			ItalianCity.Index((cList, response) {
				assert(response.status==200);
				displayKnownCities(cList, fullWidthDiv, key);
			});
		}, (response) {
			print("failed to create city! ${response.statusText}!");
		});
	});
	
	buttonDiv.nodes.add(button);
	form.nodes.add(buttonDiv);
	
	inputDiv.nodes.add(form);
	
	//ok, we are now ready to dislpay the data we loaded from the server
	displayKnownCities(cities, fullWidthDiv, key);
}

HeadingElement makeTitling(String text) {
	HeadingElement h3=new HeadingElement.h3();
	h3.text=text;
	h3.classes.add('span12'); //twitter bootstrap
	return h3;
}

DivElement makeInputRow(String label) {
	
	DivElement row = new DivElement();
	row.classes.add("control-group");//twitter bootstrap
	
	String id=label.toLowerCase().replaceAll(" ","");
	
	Element labelElement = new Element.tag('label');
	labelElement.text=label;
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
	if (!parent.nodes.isEmpty) {
		parent.nodes.clear();
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
		titleOrInput(left,city, true);

		//text about the city is a dl
		DListElement dl = new DListElement();
		dl.classes.add("dl-horizontal"); //twitter bootstrap
		
		//put in textual facts we got from server
		Element dt = new Element.tag("dt");
		dt.text="Population";
		Element dd = new Element.tag("dd");
		dd.text="${city.Population}";
		dl.nodes.add(dt);
		dl.nodes.add(dd);
		
		dt=new Element.tag("dt");
		dt.text="Province";
		dd = new Element.tag("dd");
		dd.text="${city.Province}";
		dl.nodes.add(dt);
		dl.nodes.add(dd);
		
		//put the list in place
		left.nodes.add(dl);
		
		//put a delete button there
		Element button = new Element.tag("button");
		button.text="Delete";
		button.classes.add("btn"); //twitter bootsrap
		button.classes.add("btn-danger"); //twitter bootsrap
		button.classes.add("offset1");
		button.on.click.add((Event event) {
			ItalianCity.Delete(city.Id, (dead, response) {
				assert(dead.Id==city.Id); //killed the right one, we hope...
				//delete was successful, so ask for new index of all cities
				ItalianCity.Index((newCityList, response) {
					assert(response.status==200);
					displayKnownCities(newCityList, parent, key);
				});
			});
		});
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

void titleOrInput(DivElement parent, ItalianCity city, bool title) {
	Element oldElement, newElement, input;
	
	if (title) {
		oldElement = parent.query("form");
		HeadingElement h4 =new HeadingElement.h4();
		h4.text=city.Name;
		h4.on.click.add((Event event){
			titleOrInput(parent, city, false);
		});
		newElement = h4;
	} else {
		Element form = new Element.tag("form");
		form.classes.add("form-inline"); //twitter bootstrap
		form.attributes['onSubmit']="return false;"; //we're doing the love in DART
		
		input=new Element.tag("input");
		input.attributes['type']='text';
		input.classes.add(".inputh4");
		//watch for change event
		input.on.change.add((Event event) {
			changeCityName(parent, city, event.target.value);
		});
		//escape key gets out edit mode
		input.on.keyDown.add((Event event) {
			if (event.keyCode==27) {
				titleOrInput(parent,city,true);
			}
		});
		form.nodes.add(input);
		
		oldElement = parent.query("h4");
		input.attributes['placeholder']=oldElement.text;
		
		newElement = form;
	}
	
	//swap out out the elements
	if (oldElement!=null) {
		//allows us to replace at the same location in the child order of parent
		parent.nodes = parent.nodes.map((Node n){
			if (n==oldElement) {
				return newElement;
			}
			return n;
		});
	} else {
		parent.nodes.add(newElement);
	}

	if (!title) {
		input.focus();
	}
}

void changeCityName(Element parent, ItalianCity city, String newName) {
	newName=newName.trim();
	if (newName=="") {
		titleOrInput(parent, city, true);
	}
	
	//try the networking call
	city.Name = newName;
	print("city is now ${city.Name}");
	city.Put((updatedCity, response) {
		titleOrInput(parent, updatedCity, true);
	}, (response) {
		print("Failed to update city name: ${response.statusText}");
	});
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
		print("city length is ${cities.length} and expect is ${cities.length}");
		assert(cities.length==expectedSize);
		assert(result.status==200);
	};
}
