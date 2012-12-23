library gauth;
import 'package:gauth/gauth.dart';
import 'dart:html';

class Main {
	static GauthUser me;
	
	static void LoggedIn(List<GauthUser> meList, HttpRequest req) {
		me = meList[0];
		document.query("#logout").classes.remove("invisible");
	}
	static void NotLoggedIn(HttpRequest req) {
		if (req.status==403) {
			print("forbidden, not logged in");
			document.query("#login").classes.remove("invisible");
		} else {
			print("Unexpected error trying to make check for logged in: ${req.status}, ${req.responseText}");
		}
	}
	static void CheckLogin() {
		GauthUser.Index(Main.LoggedIn, Main.NotLoggedIn, null, {"self": "true"});
	}
}