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
import 'package:nullblog/src/article_div.dart';
import 'package:nullblog/src/nullblog.dart';

class TestModule extends Module {
  configure() {
    //mock out the web for this test
    bind(Window).toType(new MockWindow());
    //bind our mock network
    bind(articleResource).toType(new MockArticleResource());
    //bind our mock network
    bind(Article_div).toType(new Article_div());
  }
}

class MockArticleResource extends Mock implements articleResource {}

void setupTwoArticles(MockArticleResource mar) {
	List<article> alist = new List<Article>();
	article a104 = new article();
	a104.Id = 104;
	alist.add(a104);
	article a103 = new article();
	a103.Id = 103;
	alist.add(a103);
	mar.when(callsTo('Index')).thenReturn(alist);
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
			List<article> result = underTest.Index();
				
			//verify things did what you thought
			underTest.getLogs(callsTo('Index')).verify(happenedOnce);
			expect(2, result.length);
			expect(104, result[0].Id);
			expect(103, result[1].Id);
		});//test
	});//group

  group('articles.html', () {
		setUp(() {
			mdv.initialize();
			print("init");
			
			fake = new article();
			fake.Id = someId;
			fake.Author = name;
			fake.Content = cont;
		});
    //now get the object under test... note we do this once per test
    //so the instances don't interact with each other (by sharing a
    //the same instance of window for example).
    test('display two articles in two divs', () {
	    Article_div underTest = injector.getInstance(Article_div);
			
			underTest.created();
			underTest.obj = fake;
			underTest.notifyChange(new PropertyChangeRecord(const Symbol('obj'))); //because underTest.obj = fake;

			return new Future(() {
				expect(underTest.author, name);
				expect(underTest.content, cont);
				expect(underTest.id, someId);
			});
    });//test
	});//group
} //main