import 'dart:html';
import 'dart:async';
import 'dart:core';

import 'package:observe/observe.dart';
import 'package:mdv/mdv.dart' as mdv;

import 'package:unittest/mock.dart';
import 'package:unittest/unittest.dart';
import 'package:unittest/html_config.dart';
import 'package:dice/dice.dart';
import 'package:fancy_syntax/syntax.dart';

import 'package:nullblog/src/articlediv.dart';
import 'package:nullblog/src/articlepage.dart'; 
import 'package:nullblog/src/nullblog.dart';
import 'package:nullblog/src/uisemantics.dart';
import 'package:nullblog/src/half_html_config.dart';

class MockArticleResource extends Mock implements articleResource {}
class MockUISemantics extends Mock implements UISemantics {}

class TestModule extends Module {
  configure() {
    bind(articleResource).toType(new MockArticleResource());
    //bind our mock network

    bind(ArticleDiv).toType(new ArticleDiv());
    bind(ArticlePage).toType(new ArticlePage());
		bind(UISemantics).toType(new MockUISemantics());  
  }
}

void clearTestArea() {
	Element testArea = query('div#test-area');
	for (Element e in testArea.children) {
			testArea.children.remove(e);
	}
}
void addTemplateDefinition(String htmlCode) {
	query('body').children.add(htmlCode);
}
void addTemplateInvocation(String htmlCode) {
	query('div#test-area').children.add(htmlCode);
}


//
// ENTRY POINT FOR TEST PAGE
//
main() {
  
  Injector injector = new Injector(new TestModule());
	article fake;
	
	const String name = "John Public";
	const String cont = "lolcatz";
	const int someId = 918;

  useHalfHtmlConfiguration();
	//put in the two templates we use in this group
	addTemplateDefinition(ArticleDiv.htmlContent);
	addTemplateDefinition(ArticlePage.htmlContent);			
	
  group('articles.html', () {
		setUp(() {
			mdv.initialize();

			//remove any stuff left from previous test
			clearTestArea();
		});
    //now get the object under test... note we do this once per test
    //so the instances don't interact with each other (by sharing a
    //the same instance of window for example).
    test('test changes to model propagate to displayed values in ArticleDiv', () {
			//this the data that we are simulating coming over the network
			fake = new article();
			fake.Id = someId;
			fake.Author = name;
			fake.Content = cont;
			
			//prepare the area on page
			addTemplateInvocation(ArticleDiv.invocation);
			
			//get the object under test and bind it to the proper bit of model
			query("#invoke-article-div").model = fake;
			query("#invoke-article-div").bindingDelegate = new FancySyntax();
			
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
			addTemplateInvocation(ArticlePage.invocation);
	
			//get the object under test and bind it to the proper bit of html
			ArticlePage underTest = injector.getInstance(ArticlePage);
			query("#invoke-article-page").model = underTest;
			query("#invoke-article-page").bindingDelegate = new FancySyntax();
	
			//create network that returns empty article set
			underTest.rez.when(callsTo('index')).thenReturn(new Future.value(new List<article>()));
			
			//now try to run the code from article page, test that the right thing happens in display
			underTest.created();
			return new Future(() {
				expect(document.query("h3"), isNotNull);
				underTest.rez.getLogs(callsTo('index')).verify(happenedOnce);
			});
			
		}); //test
		test('test that the server returns 20 articles, make 20 items on the display', () {

			//prepare the area on page
			addTemplateInvocation(ArticlePage.invocation);
	
			//get the object under test and bind it to the proper bit of html
			ArticlePage underTest = injector.getInstance(ArticlePage);
			query("#invoke-article-page").model = underTest;
			query("#invoke-article-page").bindingDelegate = new FancySyntax();
	
			//create network that returns 20 articles, all the same content
			int N = 20;
			List<article> articles = new List<articles>();
			int i = 0;
			while (i<N){
				article a = new article()..
					Id=777..
					Content="yakking"..
					Author="John Smith";
					
				articles.add(a);
				i++;
			}
			underTest.rez.when(callsTo('index')).thenReturn(new Future.value(articles));
			
			//now try to run the code from article page, test that the right thing happens in display
			underTest.created();
			
			return new Future(() {
				print("${document.query('h4')}");
				expect(document.queryAll("h4").length, equals(N));
				underTest.rez.getLogs(callsTo('index')).verify(happenedOnce);
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