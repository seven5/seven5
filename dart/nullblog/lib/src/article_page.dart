import 'dart:core';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/workarounds.dart';
import 'package:dice/dice.dart';

//Page level control for the list of articles in the blog
class Article_page extends CustomElement with ObservableMixin {
 	@Inject
	articleResource rez;
	
	@observable
	List<Article_div> allArticles;
	
	//work to do based on the network
	void created() {
		super.created();
		
		rez.index().then((List<Article_div> a) {
			allArticles=a;
		})
		.catchError( (error) {
			print("error was $error");
		});
	}
}