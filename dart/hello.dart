#import("dart:json");
#import("dart:html");

class ItalianCity {
	int Id;
	string Name;
	int Population;
	string Province;
	
	static int NOT_FETCHED = -1092;
	static string findURL = "/italiancity/";
	static string indexURL = "/italiancities/";

	static List<ItalianCity> Index(successFunc, [errorFunc]) {
		HttpRequest req = new HttpRequest();
		req.open("GET", indexURL);
		req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
			if (req.status%100==2) {
				List raw = JSON.parse(req.responseText);
				List<ItalianCity> result = new List<ItalianCity>();
				for (Map json in raw) {
					ItalianCity city = new ItalianCity.fromJson(json);
					result.add(city);
				}
				successFunc(result, req);
			} else {
				if (errorFunc!=null) {
					req.on.error.add(errorFunc);
				}
			}
		});
		req.send();
	}

	void Find(int Id, successFunc, [errorFunc]) {
		HttpRequest req = new HttpRequest();
		req.open("GET", "${findURL}/${Id}");
		req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
			if (req.status%100==2) {
				copyFromJson(JSON.parse(req.responseText));
				successFunc(this, req);
			} else {
				if (errorFunc!=null) {
					errorFunc(req);
				}
			}
		});
		req.send();
	}
	
	//convenience constructor
	ItalianCity.fromJson(Map json) {
		copyFromJson(json);
	}
	
	ItalianCity() {
		this.Id= NOT_FETCHED;
	};
	
	copyFromJson(Map json) {
		this.Id = json["Id"];
		this.Name = json["Name"];
		this.Population = json["Population"];
		this.Province = json["Province"];
	}
}

main() {
	//note that Index() is a class method (static) because it referencs the entire
	//collection, not an instance
  ItalianCity.Index(dumpAll);
	
	//newly constructed objects are empty, you must use Find() to load their content
	//from the server
  ItalianCity city = new ItalianCity();
	city.Find(0,dumpCity);
	city.Find(4,dumpCity, errorOnCity);
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

errorOnCity(HttpRequest result) {
	print("ERROR! result of 'Find' (GET) was ${result.status} ${result.statusText}");
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