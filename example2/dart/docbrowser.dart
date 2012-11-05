#import("dart:html");
#import("dart:json");


String URL_INPUT = "url-input";

void main() {
	DivElement inputDiv = document.query("#inputdiv");

	//make a form for the user to enter new cities
	Element form = new Element.tag("doc-form");
	form.classes.add("form-horizontal"); //twitter bootstrap
	form.classes.add("span6"); //twitter bootstrap
	form.classes.add("offset1"); //twitter bootstrap
	form.attributes['onSubmit']="return false;"; //we're doing the love in DART

	Element legend = new Element.tag('legend');
	legend.addText('Resource Name On Local Server');
	form.nodes.add(legend);
	
	DivElement grp = new DivElement();
	grp.classes.add("control-group");//twitter bootstrap
	
	Element labelElement = new Element.tag('label');
	labelElement.addText('Resource path or URL');
	labelElement.classes.add("control-label");//twitter bootstrap
	labelElement.attributes['for']=URL_INPUT;
	
	DivElement control = new DivElement();
	control.classes.add('controls');
	
	Element input = new Element.tag('input');
	input.attributes['type']='text';
	input.attributes['id']=URL_INPUT;
	input.attributes['placeholder']="mywiretype";
	input.on.change.add((Event event) {		
		requestAPI(document.query("#${URL_INPUT}"));
	});
	control.nodes.add(input);
	
	grp.nodes.add(labelElement);
	grp.nodes.add(control);
	
	form.nodes.add(grp);
	
	inputDiv.nodes.add(form);
}

void requestAPI(Element elem) {
	String value = elem.value, apiURL=elem.value;
	
	HttpRequest req = new HttpRequest();
	if (!value.startsWith("http://")) {
		apiURL = "http://localhost:3003/generated/api/${value}";
		elem.value = apiURL;
	}
	req.open("GET", apiURL);
	req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
		if (req.status != 200) {
			print("request for api data failed!");
			print("\t server replied with ${req.statusText}");
		} else {
			displayAPI(JSON.parse(req.responseText), "/${value}/");
		}
	});
	req.send();
}

void displayAPI(Map apiMap, String path) {
	
	DivElement apiDiv = document.query("#apidiv");
	apiDiv.nodes.clear();
	
	DivElement fullWidth = new DivElement();
	fullWidth.classes.add("span12");
	HeadingElement h2 = new HeadingElement.h2();
	h2.addText(apiMap["Name"]);
	fullWidth.nodes.add(h2);
	
	apiDiv.nodes.add(fullWidth);
	
	Collection funcs = ["Index", "Find", "Post", "Put", "Delete"];
	Collection hasBody = [ false, false, true, true, false];
	Collection methods = [ "GET", "GET", "POST", "PUT", "DELETE"];
	Collection id = ["", "123", "", "123", "123"];
	int i=0;
	for (i=0; i<funcs.length;++i) {
		DivElement result = makeAPISegment(funcs[i], methods[i], "${path}${id[i]}", hasBody[i], apiMap["${funcs[i]}Doc"]);
		if (result!=null) {
			result.classes.add("row");
			apiDiv.nodes.add(result);		
		}
	}
}



DivElement makeAPISegment(String heading, String method, String path, bool hasBody, Map apiDoc) {
	
	if (apiDoc==null) {
		return null;
	}
	DivElement result = new DivElement();
	
	DivElement parent = new DivElement();
	parent.classes.add("row");
	
	Element h = new Element.tag("span");
	h.addText(heading);
	h.classes.add("offset1 span2 lead");
	parent.nodes.add(h);
	
	Element u = new Element.tag("span");
	u.addText("${method} ${path}");
	u.classes.add("span4 offset1 lead");
	parent.nodes.add(u);
	
	result.nodes.add(parent);
	
	Element dl = new Element.tag("dl");
	dl.classes.add("span8");
	
	makeDescriptionEntry("Headers", dl, apiDoc);
	makeDescriptionEntry("QueryParameters", dl, apiDoc);
	if (hasBody) {
		makeDescriptionEntry("Body", dl, apiDoc);
	}
	makeDescriptionEntry("Result", dl, apiDoc);
	
	result.nodes.add(dl);
	
	return result;
} 

void makeDescriptionEntry(String name, Element parent, Map doc) {
		
	Element dt = new Element.tag("dt");
	dt.addText(name);
	Element dd = new Element.tag("dd");
	dd.addText(doc[name]);
	
	parent.nodes.add(dt);
	parent.nodes.add(dd);
	
}