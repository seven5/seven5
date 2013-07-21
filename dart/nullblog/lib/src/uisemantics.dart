import 'dart:html';

class UISemantics {
	
	void showNoNetworkAlert() {
		document.query(".alertbox").classes.toggle("hidden");
	}
}