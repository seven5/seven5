import 'dart:async';
import 'dart:html';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';


//This is the true implementation of the code that does the work for displaying an article
class ArticleDiv extends PolymerElement with ObservableMixin {
  
	//pulled from the server
  @observable article obj;

	//these are pushed to the UI
  @observable String author;
	@observable String content;
	@observable int id;
	
  void created() {
		super.created();
    // When 'obj' changes recompute our properties appropriately.
    bindProperty(this, const Symbol('obj'), () {
			author = obj.Author;
			id = obj.Id;
      content = obj.Content;
    });
  }
}