#import("dart:json");
#import("dart:html");

/**
  * Dead simple munger that can display a quote on the monitor.
  */
class QuoteMunger implements Munger {
	
	Map<String,Object> json;
	DivElement ourParent;
	Element quoteElement;
	Element attributionElement;
	
	// constructor just shoves the parameters into local vars
	QuoteMunger(Map<String,Object> jsonData, DivElement root) {
		this.json = jsonData;
		this.ourParent = root;
	}
	
	//10 secs should be enough to read a quote
	int preferredRunTime() {
		return 10000; //in millis
	}
	
	//run just figures out if we are start or end, but usually ends up with call to interval()
	void run(int totalTime, int start, int endExclusive) {
		
	switch (start) {
		case 0: start();
			break;
		case totalTime: end();
			break;
		default: interval(start, endExclusive);
	}
	
	void start() {
		this.quoteElement = new Element.tag("h2");
		this.quoteElement.text = json['Text'];
		
		this.attributionElement = new Element.tag("h4");
		this.attributionElement.text = json['Attribution'];
		
		ourParent.nodes.add(this.quoteElement);
		ourParent.nodes.add(this.attributionElement);
	}
	
	void end() {
		ourParent.nodes.clear();
	}
	
	void interval(int start, int endExclusive) {
		print('got an interval ${start} to ${endExclusive}');
	}
}