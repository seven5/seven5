import 'package:gitvet/src/web.dart';
import 'dart:async';
import 'dart:uri';
import 'dart:html';
import 'package:nullblog/generated/nullblog.dart';
import 'package:web_ui/watcher.dart' as watchers;

//This class does the work of manipulating the UI for the articles page.
//This class has values injected into it (like $window) that need to be
//mocked out to make the tests work out without a UI.
class ArticlesImpl {
  //injected because we have a fake DOM in the test case
  Window $window;
}