import 'dart:core';
import 'dart:html';
import 'package:polymer/polymer.dart';
import 'package:observe/observe.dart';
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/uisemantics.dart';
import 'package:dice/dice.dart';

//Page level control for the list of articles in the blog
class ArticlePage extends PolymerElement with ObservableMixin {
 	@Inject
	articleResource rez;
	
	final ObservableList<article> allArticles = new ObservableList<article>();
  
	@Inject
	UISemantics sem; 
	
	
	static final String rawHtml = '''
	<template id='article-page' syntax="fancy">
      <template repeat="{{ allArticles }}">
          <template ref="article-div" bind></template>
      </template>
			<template bind if="{{ allArticles.length == 0 }}">
          <h3 class="empty-notice">"Content, we have not", says Yoda.</h3>
      </template>
  </template>''';
  
	static final Element htmlContent = new Element.html(rawHtml);
	
	static final Element invocation = new Element.html("<template id='invoke-article-page' ref='article-page' syntax='fancy' bind>");
	
	//work to do based on the network
	void created() {
		super.created();
		
		//pull the articles from the network
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