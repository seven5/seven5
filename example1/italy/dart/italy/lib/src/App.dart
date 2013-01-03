library gauth;
import '/generated/dart';
import 'dart:html';
import 'dart:json';
import 'package:web_ui/watcher.dart' as watchers;


class App {
	static List<ItalianCity> allCities = null;
	static String apiKey= null;
	static final String PUBLIC_SETTINGS_KEY_URL="/seven5/publicsetting/italy/google_maps_key";
	
	static void loadKey() {
		//this is a "raw" dart call because we don't speak rest on this URL
		HttpRequest req = new HttpRequest();
		req.open("GET", PUBLIC_SETTINGS_KEY_URL);
		req.on.load.add(function (HttpRequestProgressEvent progressEvent) {
			if (req.status != 200) {
				print("request for google maps api key failed... check environment variable?");
				print("\t server replied with ${req.statusText}");
				fn(null);
			} else {
				print("Got the google maps API key... ${req.responseText}");
				apiKey=req.responseText;
				watchers.dispatch();//update the UI
			}
		});
		req.send();		
	}
	
	static void loadCities() {
		ItalianCity.Index(citiesOk, citiesFailed);
	}
	
	static void citiesOk(List<ItalianCity> cities, HttpRequest req) {
		allCities=cities;
		watchers.dispatch();//update the UI
	}
	
	static void citiesFailed(HttpRequest req) {
		print("failed to load the list of cities! ${req.status}, ${req.responseText}")
	}
}