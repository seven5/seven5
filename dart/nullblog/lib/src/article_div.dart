import 'dart:async';
import 'dart:html';
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/workarounds.dart';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:dice/dice.dart';


//This is the true implementation of the code that does the work for displaying an article
class Article_div extends CustomElement with ObservableMixin {
  
	@Inject
	Window window;

	@Inject
	articleResource rez;

	//pulled from the server
  @observable article obj;

	//these are pushed to the UI
  @observable String author;
	@observable String content;
	@observable int id;
	
  void created() {
		super.created();
		print("early");
    // When 'obj' changes recompute our properties appropriately.
    bindProperty(this, const Symbol('obj'), () {
		  print("hi4");
			author = obj.Author;
			id = obj.Id;
      content = obj.Content;
    });
		print("Hi");
		obj = new article();
		//for now, display the same thing every time
		rez.find(0)
		.then((article retrieved) {
			print("Hi2 ${retrieved.Author}");
			obj=retrieved;
			deliverChangeRecords();
		})
		.catchError((Object error) {
			print("Hi3");
			
			print("error during call to server $error");
		});
  }
}