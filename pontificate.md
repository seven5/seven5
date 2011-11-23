# Seven5: Pontificate

<nav>
    <ul>
        <li>[Intro](index.html)</li>
        <li>[Install](install.html)</li>
        <li>[Develop](develop.html)</li>
        <li>[Pontificate](pontificate.html)</li>
    </ul>
</nav>

Here you will find snippets of text explaining why Seven5 is the cranky, [opinionated](http://gettingreal.37signals.com/ch04_Make_Opinionated_Software.php), single-minded beast that it is.  If this were Halo, Seven5 would be the father of three with the sniper rifle making life miserable for the sword wielding 12 year olds camping by the lifts.

## Convention, not configuration

Seven5 uses conventions, not configuration, to manage how your web application will work.  Because we've chosen one way to code and one way to deploy, most of the configuration jackassery goes away.  During development you will never write a configuration file.  During deployment you will only have to place your creds in predefined places and Bob's your uncle.

Conventions also allow us to stop the insanity of repeating ourselves in each layer of the stack.  So many web frameworks think that it's normal to code your data models in multiple places: ORM, API and Javascript.  Seven5 unifies the definition of data, API, events and the rest of the system and then flows that information out where necessary.

If you want flexibility, go write Apache configuration files.


## Data structures belong in RAM, not on disk

We avoid disk storage for dynamic data because disk is slow, expensive to manage, and generally more trouble than it's worth in the age of server machines with metric tons of RAM.  Also, it's a fricking spinning platter of metal.  WTF?!? 

99% of the world doesn't need relational databases and the other 1% fucks it up.  Because of snazzy ORMs that hide SQL, developers end up with relational databases simply because that's the easiest path in the web framework.

Seven5 makes a different strategy easy.  Your data model is just, well, your data model.  You use go's data structures and store them, unmodified, in a distributed flock of RAM called [memcached](http://memcached.org/) using [gobs](http://blog.golang.org/2011/03/gobs-of-data.html) because [Rob Pike](http://en.wikipedia.org/wiki/Rob_Pike) is a righteous dude.

If you're paranoid about your entire cluster crashing then rest assured, it's trivial to provision redundant memcacheds in different locales and to take snapshots.

## Design is for designers. Programming is for programmers.

There was a time when the base materials of the web were simple enough that the "front end" of a web app could be produced by people who did not understand the "back end".  This seemed like a good idea because real programmers were snobbish about crappy browsers and foisting off quirks mode junk seemed like a good idea.

Those days are dead and buried.

Today's browsers are more standardized so the pain of quirks is fading.  Today's standards are more complex and most designers would rather design than learn about CSS selectors and event binding.  Today's pages are programs and client side libraries like jQuery and YUI require honest to god programming.

Seven5 embraces this by unifying the entire stack around languages and tools made for programmers.  Let designers design and programmers program!

## Naming

The framework is called Seven5 because the originator lives in Paris, France. All the postal codes for Paris, proper, begin with 75.  Besides, names don't matter that much.

The use of the strange pronounciation of guise is because it sounds cooler. Plus, the originator lives very close to the residence (compound?) of the  [House de Guise](http://en.wikipedia.org/wiki/House_of_Guise) which is  pronounced in this way.   The English word that is spelled the same way comes from the rumor that a dis_guise_ was used by the Duc an attempt to mask his involvement in the attempted assassination of  [Gaspard de Coligny](http://en.wikipedia.org/wiki/Gaspard_de_Coligny) that lead directly to the  [St. Bartholomew's Day Massacre](http://en.wikipedia.org/wiki/St._Bartholomew's_Day_massacre). Web frameworks  may educate in many ways.

The use of the name  [Poignard](http://en.wikipedia.org/wiki/Poignard) is because sharp, narrow tools  are often needed when working with web frameworks.  Such tools, correctly  applied, can be the killing stroke whereas huge, but relatively blunt, tools like Javascript often provide less death.  More death is better.

## Why Mongrel2

### Production Reasons

* [Mongrel2](http://www.mongrel2.org) is derived from mongrel.  Both of them have extremely well tested and secure http handling code.  Both are known to perform well under high load and to pass [valgrind](http://www.valgrind.org), so they do not leak memory.  It's solid.

* Mongrel2 is friendly for deployment/operations and can be easily configured to work in a cluster.  Mongrel2 can also handle having clusters, not necessarily in the same configuration, that handle the requests for one or more applications deployed on the cluster.  It's scalable.

### Testing Reasons

* Test the "front door" not some other path.  In other words, the best tests  use the code path that is as close---or better yet identical too--the code path  that is used by the end-user.  Unit tests in Seven5 code through the exact same dispatching (sometimes called "routing") as a request in a production deployment, even in a clustered deployment.

* Mongrel is easy to configure and control programmatically.  Seven5  exploits this ability to allow the server to be configured based on its own conventions of how to develop a web application.  During development it  should never be necessary to touch a configuration file.  Seven5 also uses this ability to programmatically start or restart mongrel2 as needed to run the developer's web application.


# Abandon all editing all ye who enter here

Testing Javascript Code In Go
-----------------------------

I'm not really sure how to make this work in practice.  Here's a couple of examples of things I'd like to write, written in english rather than as go  code:

> Set the contents of the username field to "".  Verify that the continue button is disabled.

> Set the contents of the username field to "ian".  Verify that there is
> a drop down present. If there is a dropdown present, verify that it has 
> exactly one item in it, with the contents "iansmith."

My goal would be that unit tests could be written in go, referencing the objects  used in the DSLs for CSS and HTML and somehow have it "drive" the Poignard  code through its paces.  I have the sense that the right way to do this is to have Poignard abstract the notion of JS events slightly and allow these to be  synthesized by the test harness.  I definitely do not want some crap like Selenium or other "browser level" test harnesses.  Seven5 went to a lot of trouble for  once and only once, and the tests should benefit as well.

Easy case: changing the go code in the DSL of CSS or HTML should cause immediate problems in the source code of the tests.  This is easy  because the go compiler can check this and your IDE will tell you about it  right away.  Because of the `JSGuise`, the now changed entities should cause  the JS code to fail horribly--but not until the first time you run it.

More difficult to see is how changes in the JS code can be reflected back to the go language tests automatically.  I suppose there are two basic options:

* Go code runs the show.  Since the go code is running the tests, it must be told about what/how to access `Poignard` functions, etc.  This leads to a once and only once violation since it requires that some entities be duplicated from the Javascript world to the go world.  This is not _uberbad_ because the test code runs and checks that things are ok so if you make a mistake at least it gets caught fast.

* Write some type of Javascript analysis tool and use that to make entites, at least functions, visible to the go level. There are some grammars  [laying around](http://www.antlr.org/grammar/1153976512034/ecmascriptA3.g) that could be used to extract some things from Javascript source and make them visible to test code.  

Along this latter line, but without the analysis, perhaps there could be some  simple rules associated with Poignard code.  Roughly, "Poignard can only respond to events" and these are handled according to _blah_ convention.  Then the test code could send synthetic events from the server to the  client to drive the JS code.  Similarly, JSGuise could output some "testing functions" that could be called from test code (via a network message to the  browser from the server).

## AJAX Stuff

Mongrel2 [already has](https://gist.github.com/920729) support for the WebSockets proposal by [HyBi working group](http://tools.ietf.org/html/draft-ietf-hybi-thewebsocketprotocol-07) as well as support for flash-based socket communication with JSON, XML, and blah blah blah.  We should get this for free via our mongrel2 connection.

Seven5 needs to exploit the client/server separation carefully to allow unit tests to drive both client and server.
