import 'dart:html';
import 'dart:async';
import 'package:web_ui/web_ui.dart';
import 'package:dice/dice.dart';
import 'package:nullblog/src/articlepage.dart';
import 'package:nullblog/src/uisemantics.dart';
import 'package:nullblog/src/articlediv.dart';
import 'package:nullblog/src/nullblog.dart';  //generated
import 'package:nullblog/src/workarounds.dart';  //because we can't mock fields yet
import 'package:mdv/mdv.dart' as mdv;

//setup dependencies for this page
class ArticleModule extends Module {
  configure() {
    //these two bindings are "Trivial" but really indicate that we are using the "real" implementations
		//not the mock machinery used in testing the code code Article_div
    bind(articleResource).toType(articleResource); 
		bind(ArticlePage).toType(ArticlePage);  
		bind(UISemantics).toType(UISemantics);  
    
    //we bind Document to the "real" DOM document because we are really
    //running in the browser... the use of MyWindow is temporary to work
    //around a problem with mocks
    bind(Window).toInstance(new MyWindow(window));

  }
}

ArticlePage getInjectedPage() {
	Injector injector = new Injector(new ArticleModule());
	return injector.getInstance(ArticlePage);
}




//get the object needed to control this page
void main() {
	mdv.initialize();
  /*
	runAsync( () {
		Element host = new Element.html('<div is="article-page">');
		host.model = getInjectedPage();
		ArticlePage custom = getInjectedPage()
		..host = host
		..created();
		
		Element title = document.query("#blogtitle");
		title.parent.children.add(host);

		custom.inserted();
	});*/
}
