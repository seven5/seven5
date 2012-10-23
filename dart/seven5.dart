
#library("seven5");           // Declare that this is a library.
#import("dart:json");
#import("dart:html");

class Seven5Support {
	static const int NOT_FETCHED = -1092; //signal value for object is not loaded from server
	
	//compute a URL for this call, including query params
	static encodeURL(String url, Map qp) {
		if (qp==null) {
			return url;
		}
		StringBuffer buff = new StringBuffer();
		for (var k in qp.getKeys()){
			buff.add("${k}=${qp[k]}");
		}
		//no sense trying to get to fancy as this will be url encoded anyway
		return "${url}?${buff.toString()}";
	}
	
	//typeless implementation of loading list of objects from json
	static Index(String indexURL, createList, createInstance, successFunc, errorFunc, 
		headers, params){
		HttpRequest req = new HttpRequest();
		req.open("GET", encodeURL(indexURL, params));
		if (headers!=null) {
			for (var k in headers.getKeys()){
				req.setRequestHeader(k,headers[k]);
			}
		}
		req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
			if (req.status/100==2) {
				List raw = JSON.parse(req.responseText);
				List result = createList();
				for (Map json in raw) {
					result.add(createInstance().copyFromJson(json));
				}
				if (successFunc!=null) {
					successFunc(result, req);
				}
			} else {
				if (errorFunc!=null) {
					errorFunc(req);
				}
			}
		});
		req.send();
	}
	//typeless implementation of loading single object from json
	static Find(id, findURL, obj, successFunc, errorFunc, headers, params){
		HttpRequest req = new HttpRequest();
		req.open("GET", encodeURL("${findURL}${id}", params));
		if (headers!=null) {
			for (var k in headers.getKeys()){
				req.setRequestHeader(k,headers[k]);
			}
		}
		req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
			if (req.status/100==2) {
				obj.copyFromJson(JSON.parse(req.responseText));
				successFunc(obj, req);
			} else {
				if (errorFunc!=null) {
					errorFunc(req);
				}
			}
		});
		req.send();
	}
}