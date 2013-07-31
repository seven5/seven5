--- 
layout: book
chapter: Testing The Viewer
---

### Goal: Testing a client-side app
At the end of this chapter, you should understand how to test the client side of a _Seven5_ application. You should understand how to test for the "strange cases" that can occur in web development, such as a network failure.  For this chapter, you can continue to use the code you checked out in a previous chapter, based on the branch "code-book-1". 

### Theory: The crux of the matter
This chapter is the most important of this section.  The author notes that more effort has been spent on this chapter--including the underlying technology--than any other chapter in the first section.  If one leaves out deployment, it may be fair to say that the effort to "get testing right" surpasses that spent on all the other chapters in section 1 combined.  Perhaps it is because testing, and also deployment, are so often neglected by other tools and toolchains that so much effort was expended to make them "feel good" in _Seven5_.

### Theory: Stopping the murder of client-side tests
Ample experience by this author and many others, has shown that client-side testing--testing the UI--is a nasty problem. It is "nasty" because it is so hard to keep developers testing the user interface to the level of quality needed to keep it even close to parity with the back-end.  Most developers would feel guilt, shame, and disgust if the back-end of their systems were tested as poorly as the front-end.

The front-end tests, the "UI" tests, are quickly "killed" and no longer run, if they were "born" at all, for four primary reasons.  These reasons are _all_ related to developer difficulty and hassle, not any intrinsic feature of the front-end or its technology:

* The tests are slow. As systems grow, tests which take too long end up not being run.  Even if they are run "some of the time" this reduces their value substantially as a safety net to making changes.  Test times for getting good developer feedback should be on the scale of seconds.  Tests that take hours are close to useless.

* The tests have arbitrary timing constants in them.  Most client side testing tools, notably Selenium, end up requiring the test developer to create arbitrary times to "wait" for some event to occur.  For example, "wait 100 milliseconds after pushing this button to see if a dialog box appears, if not the test fails."   Some tools are smart enough to "wait _up to_ 100 milliseconds" so at least the success cases are somewhat faster.  This has two horrific side effects.  One, it contributes to the slowness of the tests, our previous point, since the test developer typically over-provisions the waiting time to insure the test doesn't "cry wolf" simply because the timeout is too small.  Second, it makes the tests fragile.  Tests that pass on particular server might not on a developer workstation, yielding a false positive that must be investigated--and usually turned off after being found to be a false positive.  Note that tuning the time constants to minimize the test suite run time makes the fragility problem worse.

>>>> A common problem with timeouts can be seen with [continuous integration](https://en.wikipedia.org/wiki/Continuous_integration) servers (CI servers).  A suite of front end tests passes and gets used regularly until the CI server begins having false positives, at which point the tests are removed or turned off.  When there are arbitrary timing constants in the test code, it is common for tests to fail seemingly randomly due to load on the CI server.

* The tests have complex setup requirements and require back-end systems to be in a particular state to test that the front-end works correctly. The trivial example of this is requiring that a back-end server must be "up" to test the front-end. This simple requirement alone can present a problem in many circumstances because *where* the server should run, on what port, and with what data may be difficult to automate.  This is dramatically compounded by more realistic cases where the developer wants to test that UI behaves correctly in cases such as "unexpected" responses from the server, server failure, attempts to thwart system security, etc.  Such tests have a strange habit of never being written in the first place.  

* The tests must deal with asynchrony.  Modern web UIs are deeply asynchronous--put another way, they are single-threaded.  They require the client-side code to "hold state" and then process callbacks based on changes.  Pushing a button on the UI triggers a callback; receiving a response from a previous network call triggers a callback.  Most unit tests on the server-side of a system are _synchronous_ in that they put the system into a state and directly test its behavior in that state.  With a web-based system, the client side must be requested to reach state X and when the state is reached and the requesting code notified, the system's behavior is tested.  This type of test design is perhaps more difficult, but certainly unusual, for developers that are expecting the synchronous-style of testing.

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

As we explained previously, `ArticlePage` is responsible for the large scale control of the page behavior; it is the object under test here.  In this case we are testing that if the server were to return 0 articles, which is possible in the abstract but not with our existing implementation of `runnullblog`, we get a sensible UI telling the user that there is no content.

Critical to the construction of this test is the line:

```
underTest.rez.when(callsTo('index')).thenReturn(new Future.value(new List<article>()));
```

This line is using a [mock object](http://en.wikipedia.org/wiki/Mock_object) for the implementation of the network resource, `rez`.  The [mocking framework]


