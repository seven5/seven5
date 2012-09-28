#import("dart:json");
#import("dart:html");
#import("modena")

main() {
	String url = "http://localhost:3003/quote/";
  HttpRequest request = new HttpRequest.get(url, onSuccess);
}
