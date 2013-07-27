import 'dart:core';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/workarounds.dart';
import 'package:nullblog/src/uisemantics.dart';
import 'package:dice/dice.dart';
import 'package:stack_trace/stack_trace.dart';

//Page level control for the list of articles in the blog
class ArticlePage extends PolymerElement with ObservableMixin {
 	@Inject
	articleResource rez;
	
	final ObservableList<article> allArticles = new ObservableList<article>();
  
	@Inject
	UISemantics sem; 
	
	bool get isEmpty {
		print("called isEmpty? ${allArticles.isEmpty}");
		return allArticles.isEmpty;
	}
	
	//work to do based on the network
	void created() {
		super.created();

    new ListPathObserver(allArticles, 'isEmpty').changes.listen((changeUpdate) {
			print("list path observer: here");
			//this extra notification is because isEmpty is a property that is derived from allArticles
			//notifyChange(new PropertyChangeRecord(const Symbol('isEmpty'))); 
		});
		
		//print("created: ${allArticles.isEmpty}");
		
		rez.index().then((List<article> a) {
			allArticles.clear();
			allArticles.addAll(a);
			//if you add throw Exception("foo"); here you can trigger the error message for no network below
		})
		.catchError( (error) {
			//somewhat naive, this assumes that any error is network related
			print("ArticlePage: received an error trying to get articles from network: $error");
			sem.showNoNetworkAlert();
		});
	}
}