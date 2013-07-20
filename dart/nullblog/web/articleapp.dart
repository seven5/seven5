import 'dart:html';
import 'dart:async';
import 'package:web_ui/web_ui.dart';
import 'package:dice/dice.dart';
import 'package:nullblog/src/article_page.dart';
import 'package:nullblog/src/article_div.dart';
import 'package:nullblog/src/nullblog.dart';  //generated
import 'package:nullblog/src/workarounds.dart';  //because we can't mock fields yet
import 'package:mdv/mdv.dart' as mdv;

//setup dependencies for this page
class ArticleModule extends Module {
  configure() {
    //these two bindings are "Trivial" but really indicate that we are using the "real" implementations
		//not the mock machinery used in testing the code code Article_div
    bind(articleResource).toType(articleResource); 
		bind(Article_page).toType(Article_page);  
    
    //we bind Document to the "real" DOM document because we are really
    //running in the browser... the use of MyWindow is temporary to work
    //around a problem with mocks
    bind(Window).toInstance(new MyWindow(window));

  }
}

//get the object needed to control this page
void main() {
	mdv.initialize();
	
	Injector injector = new Injector(new ArticleModule());
	
	Article_page article_page = injector.getInstance(Article_page);
  article_page.host = new DivElement();
  
  /*var lifecycleCaller = new ComponentItem(article_page)..create();*/
	article_page.created();
	document.body.nodes.add(article_page.host);
  /*lifecycleCaller.insert();*/
	article_page.inserted();
	
}

