import 'dart:html';
import 'dart:uri';
import 'dart:async';
import 'package:dice/dice.dart';
import 'package:web_ui/web_ui.dart';
import 'package:web_ui/watcher.dart' as watchers;
import 'package:nullblog/src/articlesimpl.dart';
import 'package:nullblog/src/web.dart';
import 'package:nullblog/generated/nullblog.dart';

//setup dependencies for this page
class ArticlesModule extends Module {
  configure() {
    //real implementation of the network
    bind(ArticleResource).toType(new ArticleResource());
    
    //we bind Document to the "real" DOM document because we are really
    //running in the browser... the use of MyWindow is temporary to work
    //around a problem with mocks
    bind(Window).toInstance(new MyWindow(window));

    //object for this page
    bind(ArticlesImpl).toType(new ArticlesImpl());
  }
}

//this controls showing anything the page
bool pageReady = false;

//this is object used for this page
ArticlesImpl impl;

//get the object needed to control this page
void main() {
  var injector = new Injector(new ArticlesModule());
  injector.getInstance(ArticlesImpl).then((ArticlesImpl hasBeenInjected) {
    impl = hasBeenInjected;
    pageReady=true;
    watchers.dispatch();
  });
}
