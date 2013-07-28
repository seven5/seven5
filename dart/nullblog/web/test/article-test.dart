import 'dart:html';
import 'dart:async';
import 'dart:core';

import 'package:observe/observe.dart';
import 'package:mdv/mdv.dart' as mdv;

import 'package:unittest/mock.dart';
import 'package:unittest/unittest.dart';
import 'package:unittest/html_config.dart';
import 'package:dice/dice.dart';

import 'package:nullblog/src/webmocks.dart';  //for tests to avoid the browser
import 'package:nullblog/src/workarounds.dart'; //workaround for bad mock framework
import 'package:nullblog/src/articlediv.dart';

import 'package:nullblog/src/articlepage.dart'; //compiled version
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/uisemantics.dart';

class TestModule extends Module {
  configure() {
    //mock out the web for this test
    //bind(Window).toType(new MockWindow());
    //bind our mock network
    bind(articleResource).toType(new MockArticleResource());
    //bind our mock network
    bind(ArticleDiv).toType(new ArticleDiv());
    bind(ArticlePage).toType(new ArticlePage());
		bind(UISemantics).toType(new MockUISemantics());  
  }
}

class MockArticleResource extends Mock implements articleResource {}
class MockUISemantics extends Mock implements UISemantics {}

void setupTwoArticles(MockArticleResource mar) {
	List<article> alist = new List<Article>();
	article a104 = new article();
	a104.Id = 104;
	alist.add(a104);
	article a103 = new article();
	a103.Id = 103;
	alist.add(a103);
	mar.when(callsTo('index')).thenReturn(alist);
}

void setTestTemplate(templateBody, templateInvocation){
	Element testArea = document.query("#test-area");
	query('body').children.add(templateBody);
	testArea.children.clear();
	testArea.children.add(templateInvocation);
}

//
// ENTRY POINT FOR TEST PAGE
//
main() {
  useHtmlConfiguration();
  
  Injector injector = new Injector(new TestModule());
	article fake;
	
	const String name = "John Public";
	const String cont = "lolcatz";
	const int someId = 918;

  group('sanity check', () {
		test('prove that the setup works', () {
	  	articleResource underTest = injector.getInstance(articleResource);
			//prepare mocks
			setupTwoArticles(underTest);
				
			//run test... this is "real" call to get the list of articles
			List<article> result = underTest.index();
				
			//verify things did what you thought
			underTest.getLogs(callsTo('index')).verify(happenedOnce);
			expect(2, result.length);
			expect(104, result[0].Id);
			expect(103, result[1].Id);
		});//test
	});//group

  group('articles.html', () {
		setUp(() {
			mdv.initialize();

			fake = new article();
			fake.Id = someId;
			fake.Author = name;
			fake.Content = cont;
			
		});
    //now get the object under test... note we do this once per test
    //so the instances don't interact with each other (by sharing a
    //the same instance of window for example).
    test('test changes to model propagate to displayed values in ArticleDiv', () {
			//prepare the area on page
	    setTestTemplate(ArticleDiv.htmlContent, ArticleDiv.invocation);
	
			//get the object under test and bind it to the proper bit of html
	    //ArticleDiv underTest = injector.getInstance(ArticleDiv);
			query("#invoke-article-div").model = fake;
			//underTest.obj = fake;

			return new Future(() {
				//check structure of template instantiation
				expect(document.query('p.lead'), isNotNull);
				expect(document.query('p.garbage'), isNull); //sanity
				expect(document.query('h4.author'), isNotNull);
				expect(document.query('div.article-div-main'), isNotNull);

				//check content
				expect(document.query('h4.author').text.contains(name), isTrue);
				expect(document.query('p.lead').text.contains(cont), isTrue);
			});
    });//test
    test('test that the server returns 0 articles, we do something sensible', () {

			//prepare the area on page
	    setTestTemplate(ArticlePage.htmlContent, ArticlePage.invocation);
	
			//get the object under test and bind it to the proper bit of html
			ArticlePage underTest = injector.getInstance(ArticlePage);
			query("#invoke-article-page").model = underTest;
	
			//create network that returns empty article set
			underTest.rez.when(callsTo('index')).thenReturn(new Future.value(new List<article>()));
			
			//now try to run the code from article page, test that the right thing happens in display
			underTest.created();
			
			return new Future(() {
				expect(document.query("h3.empty-notice"), isNotNull);
			});
	
			
		}); //test
    test('test that if network fails we display an error', () {
	    ArticlePage underTest = injector.getInstance(ArticlePage);
	
			//create a failing network
			underTest.rez.when(callsTo('index')).thenReturn(new Future.error('you lose'));

			//now try to run the code from article page to see what it does
			underTest.created();
			
			//with a bad network, verify we showed the alert UI
			return new Future(() {
				underTest.sem.getLogs(callsTo('showNoNetworkAlert')).verify(happenedOnce);
			});
		
		}); //test
	});//group
} //main