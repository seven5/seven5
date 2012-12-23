library seven5support;

import 'dart:json';
import 'dart:html';

class Seven5Support {
	static const int NOT_FETCHED = -1092; //signal value for object is not loaded from server
	
	//compute a URL for this call, including query params
	static encodeURL(String url, Map qp) {
		if (qp==null) {
			return url;
		}
		StringBuffer buff = new StringBuffer();
		qp.forEach((String k,String v) {
			buff.add("${k}=${qp[k]}");
		});
		//no sense trying to get to fancy as this will be url encoded anyway
		return "${url}?${buff.toString()}";
	}
	
	static void addHeaders(Map headers, HttpRequest req) {
		if (headers!=null) {
			headers.forEach((String k,String v) {
				req.setRequestHeader(k,v);
			});
		}
	}
	
	static void Index(String indexURL, Function createList, Function createInstance, Function successFunc, Function errorFunc, 
		Map headers, Map params){
		HttpRequest req = new HttpRequest();
		
		req.open("GET", encodeURL(indexURL, params));
		
		Seven5Support.addHeaders(headers,req);
		
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
	//resourceCallWithObjectResult is called to make a call on the resource and pass the result through to the 
	//success function as parameter.  POST, PUT, DELETE, and FIND all use this code because they all expect a
	//single object as the result of their call (in the success case).
	static void resourceCallWithObjectResult(String method, String encodedURL, dynamic obj, Function successFunc, 
		Function errorFunc, Map headers, String body){
		HttpRequest req = new HttpRequest();
		req.open(method, encodedURL);
		
		Seven5Support.addHeaders(headers,req);
		
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
		req.send(body);
	}
	//singleInstance is used by PUT, DELETE, and FIND because they _address_ a particular object as well as
	//expecting as a single object as a return value.
	static void singleInstance(String method, int id, String resURL, dynamic obj, Function successFunc, Function errorFunc,
		Map headers, Map params, String bodyContent) { 
		Seven5Support.resourceCallWithObjectResult(method, encodeURL("${resURL}${id}", params), obj, successFunc, errorFunc, 
			headers, bodyContent);
	}
	
	static void Put(String bodyContent, int id, String resURL, dynamic obj, Function successFunc, Function errorFunc, 
		Map headers, Map params){
			Seven5Support.singleInstance("PUT", id, resURL, obj, successFunc, errorFunc, headers, params, bodyContent);
	}
	static void Delete(int id, String resURL, dynamic obj, Function successFunc, Function errorFunc, 
		Map headers, Map params){
			Seven5Support.singleInstance("DELETE", id, resURL, obj, successFunc, errorFunc, headers, params, null);
	}
	static void Find(int id, String resURL, dynamic obj, Function successFunc, Function errorFunc, 
		Map headers, Map params){
			Seven5Support.singleInstance("GET", id, resURL, obj, successFunc, errorFunc, headers, params, null);
	}
	
	static void Post(String bodyContent, String resURL, dynamic obj, Function successFunc, Function errorFunc, 
		Map headers, Map params){
			Seven5Support.resourceCallWithObjectResult("POST", encodeURL("${resURL}", params), obj, successFunc, 
			errorFunc, headers, bodyContent);
	}
}