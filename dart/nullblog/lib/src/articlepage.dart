import 'dart:core';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/workarounds.dart';
import 'package:nullblog/src/uisemantics.dart';
import 'package:dice/dice.dart';

//Page level control for the list of articles in the blog
class ArticlePage extends PolymerElement with ObservableMixin {
 	@Inject
	articleResource rez;
	
	final ObservableList<article> allArticles = new ObservableList<article>();
  
	@Inject
	UISemantics sem; 
	
	//work to do based on the network
	void created() {
		super.created();

		rez.index().then((List<article> a) {
			allArticles.clear();
			allArticles.addAll(a);
			//(const Symbol('allArticles'), null, a);
		})
		.catchError( (Error error) {
			//somewhat naive, this assumes that any error is network related
			print("error is $error");
			sem.showNoNetworkAlert();
		});
	}
}