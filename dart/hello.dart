#import("dart:json");
#import("dart:html");

class ItalianCity {
	int Id;
	string Name;
	int Population;
	string Province

	static List<ItalianCity> Index(successFunc) {
		string resource = "/italiancities/";
		new HttpRequest.get(resource, function (HttpRequest request) {
			List <ItalianCity> result = JSON.parse(req.responseText);
			successFunc(result);
		});
	}
}

dump(List<ItalianCity> cities) {
	for (ItalianCity city in cities) {
		print city.Name
	}
}

main() {
  ItalianCity.Index(dump);
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