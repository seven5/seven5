--- 
layout: book
chapter: How The Blog Viewer Works
---

### Goal: Understand a client-side app
At the end of this chapter, you should understand the basics of how a page's data is retrieved and how the page is laid out in a _Seven5_ user interface. You should also understand how the "model" and the "view" are related to the display of the web interface.  For this chapter, you can continue to use the code you checked out in a previous chapter, based on the branch "code-book-1". 

### Theory: Dependency injection
The topic of [Dependency injection](http://en.wikipedia.org/wiki/Dependency_injection) is too large to be covered well here.  For the time-being, we will use relatively simplistic notion that dependency injection is simply a way to allow two code paths to share a single implementation by each supplying some needed values--the _dependencies_.  In particular for our purposes, we allow the "production" code path and the "test" code path to share the code defined in the "private source" of the application.  In our example, this code lives in `/tmp/book/nullblog/dart/nullblog/lib/src`.

### Pratice: Dice and the app file
The file `articleapp.dart`, in the "web directory" or `/tmp/book/nullblog/dart/nullblog/web`, makes heavy use of [Dice](http://pub.dartlang.org/packages/dice), by Lars Tackmann, a dependency injector for the Dart language.  At the top of the "app file" is this bit of configuration code:

```
class ArticleModule extends Module {
  configure() {
    bind(articleResource).toType(articleResource); 
	bind(ArticlePage).toType(ArticlePage);  
	bind(UISemantics).toType(UISemantics);  
  }
}

ArticlePage getInjectedPage() {
	Injector injector = new Injector(new ArticleModule());
	return injector.getInstance(ArticlePage);
}

```

The [details of Modules in Dice](https://github.com/ltackmann/dice) are described elsewhere.  For our purpose, this is mechanism for specifying a mapping from a type name, such as `articleResource` above, to an implementation class, such as `articleResource` above!  This type of configuration is actually desirable because this is the production code path--the implementation of the machine generated implementation of `articleResource` is the one expected. 

The key object for `article.html` is an implementation of `ArticlePage`.  This is retrieved in the method `getInjectedPage` method above.  The object that has been injected can be equally well thought of as being configured by the `configure` nested method in `ArticleModule`.

### Practice: The app main
```
void main() {
	mdv.initialize();
	TemplateElement.syntax['fancy'] = new FancySyntax();
  
	ArticlePage p = getInjectedPage();
	p.created();
	
	query("body").children.add(ArticleDiv.htmlContent);
	query('body').children.add(ArticlePage.htmlContent);
	
	query('.content-column').children.add(ArticlePage.invocation);
	
	query("#invoke-article-page").model = p;
}
```

The application `main()` shown above for `articles.html` is almost entirely configuration of the application--it really has no functionality.  This is both normal and desirable.   The first two sections are boiler-plate code needed to initialize the [model driven views](http://www.polymer-project.org/platform/mdv.html) framework and creates an instance of `ArticlePage` (see above).   The following two lines add templates to the page (see below) that specify the visual display (HTML code) for the `ArticleDiv` and `ArticlePage` classes that this page uses.  The particular display code used should *not* be in this file: the display code will need to be tested and code in this file is _not_ visible to tests!  The two templates are added to the "body" of the HTML page (via `query('body').children.add`) but since they only _define_ templates, and result in no HTML code output, their location is not terribly important.  

The next line is critical, it _instantiates_ the sole `ArticlePage` template that controls the entire page and places it into the proper place in the DOM.  This has the effect of _adding_, perhaps even better "generating", many DOM nodes into the page.  Exactly what nodes are generated can vary based on the particulars of the template `ArticlePage.htmlContent` and what model is bound to the template.

The final line binds the _model_ to the `ArticlePage` template.  For convenience, we keep the model and the template definition in the same file, in this case `/tmp/book/nullblog/dart/nullblog/lib/src/ArticlePage.dart`.  The "model" exposed by `ArticlePage` can be, roughly, thought of as any fields that are marked with the annotation `@observable` in the source code of articlepage.dart.

### Theory: Templates
User interfaces is _Seven5_ are implemented chiefly through the notion of ["templates"](http://www.dartlang.org/docs/tutorials/templates/).  These correspond to snippets of HTML code, stored in a web page, that can be instantiated to create repeated blobs of HTML.  A single template may be instantiated more than once.  The template `article-div` has been instantiated once for each article in the screen shot shown in the previous chapter.  The custom is to use the field names `rawHtml` and `htmlContent` for the template's raw text and DOM representation, respectively, inside a Dart class.

### Practice: The page file
Page files, Dart code files suffixed with "page.dart" for a given noun in a REST API, are responsible for the large-scale structure of the resulting interface.

```
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
```

The above is the declarative portion of the code for `ArticlePage` ("the model" and "the template").  The lines marked with the annotation `@Inject` indicate that these are created via the dependency injection mechanism explained earlier in this chapter.   The precise run-time type of the fields "rez" or "sem" are controlled by the `configure()` section of the code above and are outside the control of this class.  Code inside `ArticlePage` can assume that the class, in an object-oriented sense, is *at least* the declared type such as `articleResource` for `rez`.

The code above also declares a field called `allArticles` that is of type `ObservableList<article>`.  This is the *crucial connection* in this application.  This list of articles will be populated with values of type `article` that were born, on the server, as the Go type `ArticleWire`.  They will be retrieved via the client-side manifestation of the article "resource". 
    
By reading the HTML code inside the `rawHtml` declaration, one can see that templates may repeat and that some portion of this template is repeated once for each element in `allArticles`.  The repeated portion is, of course, itself a call to instantiate a template, in this case the `article-div`! As is normal for a web application, the template includes logic to create some HTML code for handling the case where there are no articles--something that was not shown in our previous screenshot, but is still important.

### Theory: Template naming
The use of `article-div` to represent the template defined in the dart class `ArticleDiv.dart` is certainly questionable.  The implied rule here is to use camel case names that begin with uppercase letters for Dart types, but dash-separated, all lowercase names for their corresponding HTML templates.  (This also implies that the query to find the template for `ArticleDiv` is `query("#article-div")` in Dart code, also somewhat strange.)  This naming convention was chosen to preserve the "traditional" way of naming entities in their respective languages but makes automatic generation of various names more problematic.

### Practice: Transferring Data
In `ArticlePage`, the critical code to read the article content is contained in the function `created()` and makes heavy use of Dart futures:

```
void created() {
	super.created();
	
	//pull the articles from the network
	rez.index().then((List<article> a) {
		allArticles.clear();
		allArticles.addAll(a);
	})
	.catchError( (error) {
		//somewhat naive, this assumes that any error is network related
		sem.showNoNetworkAlert();
	});
}
```

Simplifying a small amount, this code requests the list of articles from the server side implementation of nullblog via the `index` method on `rez`.  On the server, this corresponds to the call of the same name on the type `ArticleResource`.  There are two cases that must be dealt with in terms of the server's response to our request: the normal case and the error case.  The normal case is inside the `then` section of the future.  In the normal case, all the articles found (in the parameter `a`) are placed into the (model) field `allArticles`.  In the error case, the `UISemantics` object is engaged (`sem`) to display an alert to the user.

The code above updates the model field `allArticles`.  Changes to the model fields marked with the annotation `@observable` or certain types, such as `ObservableList` in the case of `allArticles` are automatically propagated to the user interface.  There is no need to "force a redraw" as one might expect in other toolkits.

It is not a coincidence that the two most critical "external" classes for this critical portion of the code are ultimately controlled by the dependency injector, Dice.  This will be shown in the next chapter to be critical for testing.

### Theory: Model-driven views
There are numerous scholarly articles about model-driven views and their intellectual predecessor, the model view controller architecture, MVC (Reenskaug 1979).  While these are of interest to us as historical documents--"where did we come from" stories--more directly relevant to this book are the [polymer project](http://www.polymer-project.org/) and its sibling, [Polymer.dart](http://pub.dartlang.org/packages/polymer).  The latter is principally a port of the former to the Dart language, as the former is targeted at Javascript.

The basic capability offered by Model Driven Views, or MDV, is to bind an object, in our case a Dart class, to a template and have variable values "flow" from the model into the template.  This can be seen in the code for `ArticlePage` in the template text of `rawHtml` with the [expressions](https://github.com/dart-lang/fancy-syntax#syntax) inside mustaches ({}).  Although it has not been shown in this book so far, the "observable" expressions in the template are updated automatically as the values in the model change. In a future chapter we will give an example where data flows the opposite direction, from HTML to Dart, equally easily.

Ignoring the issue of the Dart programming language more broadly, the decision was made to use a particular subset of MDV, a part of Polymer, as the front-end technology for _Seven5_. This was based primarily on two factors:

* The ability to "bind" a particular model to a template programmatically, or perhaps "by hand", offers significant testing advantages.  Any approach where this binding is done outside programmer control makes testing harder or impossible as it typically prevents the use of a dependency injector.  This is a trade-off that was viewed as desirable where additional, boilerplate code has to typically be added to the production code path in return for the ability to test more conveniently.

* As was stated in the previous chapter, there is large developer advantage in allowing a simple page "reload", one keystroke, to build and re-run all the client-side code.  This desire disallowed various compiler-type tools.

A less decisive, but still relevant, reason to use model-driven views was the ability to exploit modern browser features without compromising on browser compatibility.  The Polymer project uses "polyfills", sometimes called "work-alikes", to emulate the features of modern browsers on older browsers that do not support the particular feature.  For example, the use of templates in the code above is "supported" in most any browser through the use of polyfills.  The  polyfills idea, combined with compiling Dart to Javascript explained later, allows, in the author's opinion, the most civilized client-side programming environment to date that runs on any reasonable web browser.








    
    








