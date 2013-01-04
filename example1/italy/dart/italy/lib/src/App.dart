library gauth;
import '/generated/dart';
import 'dart:html';
import 'dart:json';
import 'package:web_ui/watcher.dart' as watchers;


class App {
	static List<ItalianCity> allCities = null;

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