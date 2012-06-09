#library('seven5');
#import("dart:html");
#import("dart:json");
#import("jsonobject.dart");


class RestService {
  Map<String,Object> all(String url) {
    XMLHttpRequest request = new XMLHttpRequest();
    request.open("GET", url);
    request.on.readyStateChange.add((Event e){
      if (request.readyState == XMLHttpRequest.DONE &&
          (request.status == 200 || request.status == 0)) {
        onSuccess(request); // called when the GET completes
      }
      request.send();
    });
    return new Map<String,Object>();
  }
  
  void onSuccess(XMLHttpRequest request) {
    print(request);
    print("here");
  }
}
