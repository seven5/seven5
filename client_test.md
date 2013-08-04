--- 
layout: book
chapter: Testing The Viewer
---

### Goal: Testing a client-side app
At the end of this chapter, you should understand how to test the client side of a _Seven5_ application. You should understand how to test for the "strange cases" that can occur in web development, such as a network failure.  For this chapter, you can continue to use the code you checked out in a previous chapter, based on the branch "code-book-1". 

### Theory: The crux of the matter
This chapter is the most important of this section.  The author notes that more effort has been spent on this chapter--including the underlying technology--than any other chapter in the first section.  If one leaves out deployment, it may be fair to say that the effort to "get testing right" surpasses that spent on all the other chapters in section 1 combined.  Perhaps it is because testing, and also deployment, are so often neglected by other tools and toolchains that so much effort was expended to make them "feel good" in _Seven5_.

### Theory: Stopping the murder of client-side tests
Ample experience by this author and many others has shown that client-side testing--testing the UI--is a nasty problem. It is "nasty" because it is so hard to keep developers testing the user interface to the level of quality expected, say, of the same system's back-end.  Most developers would feel guilt, shame, and disgust if the back-end of their systems were tested as poorly as the front-end.

The front-end tests, the "UI" tests, are "killed" and no longer run, if they were "born" at all, for four primary reasons.  These reasons are _all_ related to developer difficulty and hassle, not any intrinsic feature of the front-end or its technology:

* The tests are slow. As systems grow, tests which take too long end up not being run.  Even if they are run "some of the time" this reduces their value substantially as a safety net for making changes.  Test times for getting good developer feedback should be on the scale of seconds.  Tests that take hours are close to useless.

* The tests have arbitrary timing constants in them.  Most client side testing tools, notably Selenium, end up requiring the test developer to create arbitrary times to "wait" for some event to occur.  For example, "wait 100 milliseconds after pushing this button to see if a dialog box appears, if not the test is deemed to have failed."   Some tools are smart enough to "wait _up to_ 100 milliseconds" so at least the success cases are somewhat faster.  This has two horrific side effects.  One, it contributes to the slowness of the tests, the previous point, since the test developer typically overprovisions the waiting time to insure the test doesn't "cry wolf" simply because the timeout is too small.  Second, it makes the tests fragile.  Tests that pass on particular server might not on a developer workstation, yielding a false positive that must be investigated--and usually turned off after being found to be a false positive.  Note that tuning the time constants to minimize the test suite run time makes the fragility problem worse.

>>>> A common problem with timeouts can be seen with [continuous integration](https://en.wikipedia.org/wiki/Continuous_integration) servers (CI servers).  A suite of front end tests passes and gets used regularly until the CI server begins having false positives, at which point the tests are removed or turned off.  When there are arbitrary timing constants in the test code, it is common for tests to fail seemingly randomly due to load variation on the CI server.

* The tests have complex setup requirements and require back-end systems to be in a particular state to test that the front-end works correctly. The trivial example of this is requiring that a back-end server must be "up" to test the front-end. This simple requirement alone can present a problem in many circumstances because *where* the server should run, on what port, and with what data may be difficult to automate.  This is dramatically compounded by more realistic cases where the developer wants to test that UI behaves correctly in cases such as "unexpected" responses from the server, server failure, attempts to thwart system security, etc.  Such tests have a strange habit of never being written in the first place.  

* The tests must deal with asynchrony.  Modern web UIs are deeply asynchronous--put another way, they are single-threaded.  They require the client-side code to "hold state" and then process callbacks based on changes.  Pushing a button on the UI triggers a callback; receiving a response from a previous network call triggers a callback.  Most unit tests on the server-side of a system are _synchronous_ in that they put the system into a state and directly test its behavior in that state.  With a web-based system, the client side must be requested to reach state X and when the state is reached and the requesting code notified, the system's behavior is tested.  This type of test design is perhaps more difficult, but certainly unusual, for developers that are expecting the synchronous style of testing.

_Seven5_ allows a developer to _easily_ write tests without any of these barriers meaning that front-end code can be as well tested as the back-end.  FTW.

### Practice: A simple test

We will begin our discuss of the practice of testing the UI with something that is not visible to users on the production, or "normal", code path.

```
test('test that if the server returns 0 articles, we do something sensible', () {

		//prepare the area on page
		addTemplateInvocation(ArticlePage.invocation);

		//get the object under test and bind it to the proper bit of html
		ArticlePage underTest = injector.getInstance(ArticlePage);
		query("#invoke-article-page").model = underTest;

		//create network that returns empty article set
		underTest.rez.when(callsTo('index')).thenReturn(new Future.value(new List<article>()));
		
		//now try to run the code from article page, test that the right thing happens in display
		underTest.created();
		return new Future(() {
			expect(document.query("h3"), isNotNull);
			underTest.rez.getLogs(callsTo('index')).verify(happenedOnce);
		});
		
	}); //test

```

As we explained previously, `ArticlePage` is responsible for the large scale control of the page behavior; it is the object under test here.  In this case we are testing that if the server were to return 0 articles, which is possible in the abstract, but not with our existing implementation of `runnullblog`, we get a sensible UI telling the user that there is no content.

Critical to the construction of this test is the line:

```
underTest.rez.when(callsTo('index')).thenReturn(new Future.value(new List<article>()));
```

This line is using a [mock object](http://en.wikipedia.org/wiki/Mock_object) for the implementation of the network resource, `rez`.  The mocking framework is supplied by Dart itself, and we use Dice to inject this mock into the `ArticlePage`.  The code path inside `ArticlePage` is identical for test and production, only the implementation of `articleResource` is different, and the called code should not be concerned about that.

### Theory: Recipe for a test

The general structure of a UI test in _Seven5_ is this:

```
    1. install invocation of templates in the test area
    
    2. get the object under test from the dependency injector
    
    3. add "expected" calls to mock objects to return the values for the test
    
    4. run the code under test
    
    5. return a Future
    
    6. inside the future, check that the UI is in the desired state
    
    7. inside the futer, check that any mock objects have been called the expected number of times, with expected parameters, etc.
```

Step 5 returns the Future that is executed (at some later point) to run the final two tests.  This type of "return some code to check in a minute" is directly supported by the unit test framework of Dart.  The reason that this _must_ be inside a Future is because the code under test in step four is likely to make "changes" to the UI that will be "happen" asynchronously.  By returning a Future the test insures that the checks occur after the update to the UI has completed.  

It should be note that this recipe does not show the templates actually being installed in the page.  It is typically more convenient to do this in a `setUp()` method that can be shared among all tests.

### Theory: Idiomatic structure

The previous section outlines the tactics for a single test.  At a higher level,  the strategy in the test code discussed above is a common one for testing websites powered by Seven5. Typically a web page maintains one or more connections to the server to perform its function.  If you follow the recipe shown with the example code and theoretical structure shown just previously, it should always be possible to simulate _any_ result from the server(s).  Because the server is being simulated via the injection of "phony" resources, it is also not necessary to have a server "up" to run these tests.

### Practice: Testing failures with Futures

Because we have a dependency injector that allows us to simulate the network's behavior, we can create tests that show that the UI works properly in the face of failures.  Here is a simple test of this simple blog viewer when the network fails:

```
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
```

Again, the critical line involves telling the mock framework to return a particular value when the method `index` is called:

```
underTest.rez.when(callsTo('index')).thenReturn(new Future.error('you lose'));
```

This test demonstrates how to use the ```Future.error``` constructor to return a Future that is always going to indicate an error.

### Practice: Test output

Below is a screen snap of test output:

![Screen Shot Of Test Output](https://www.evernote.com/shard/s238/sh/cc75c502-f2cd-475b-9d7d-868f16a33684/334226b7a4e787864348d3cdd633fa61/deep/0/web_test.png)

This output has been kept intentionally terse so large numbers of tests can be run with minimal "noise".  The links shown in the snap, with the name of each test, can be clicked on to run a single test.  When viewing a single test in this way, the "test area" that contains the rendered result of the test _is_ shown as this can be helpful in debugging.  

