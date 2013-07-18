import 'package:mdv/mdv.dart' as mdv;
import 'dart:html';
import 'dart:async';
import 'package:dice/dice.dart';
import 'package:nullblog/src/article_div.dart';
import 'package:nullblog/src/nullblog.dart';  //generated
import 'package:nullblog/src/workarounds.dart';  //because we can't mock fields yet

//setup dependencies for this page
class ArticleModule extends Module {
  configure() {
    //these two bindings are "Trivial" but really indicate that we are using the "real" implementations
		//not the mock machinery used in testing the code code Article_div
    bind(articleResource).toType(articleResource); 
		bind(Article_div).toType(Article_div);  
    
    //we bind Document to the "real" DOM document because we are really
    //running in the browser... the use of MyWindow is temporary to work
    //around a problem with mocks
    bind(Window).toInstance(new MyWindow(window));

  }
}

//get the object needed to control this page
void main() {
	mdv.initialize();
	
	/*
	Timer.run(() {
		//from here it's all ripped off from this test
		//https://github.com/dart-lang/web-ui/blob/master/test/data/input/component_created_in_code_test.html#L28
		Injector injector = new Injector(new ArticleModule());
		
		Article_div impl = injector.getInstance(Article_div);
		impl.host = new DivElement();
		
		//this should be automatic but isn't yet
		impl.created();
		//put at the end of the body
		document.body.nodes.add(impl.host);
		impl.inserted();
		
		print("$impl");
	});*/
}

