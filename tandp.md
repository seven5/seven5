--- 
layout: book
chapter: Background
---

## A few short history lessons 

In April 2010, Rob Pike gave the first public talk about Go to a class at Stanford.  In that talk, he, quite diplomatically, asserts that that many people had come to a wrong-headed conclusion about the relationship of dynamism to programming being fast and fun.  He claimed that it was, reading between the politely-phrased lines, the poor design of statically-typed languages that led to this conclusion; he grouped C++ and Java into the camp of "efficient and statically-typed" languages that ruined the fun in programming.  In that talk, he made the case that Go can be as "light on the page" and fun as languages like Ruby or Python and yet be statically-typed and efficient.  Seven5 uses Go for all server-side code.   

In October 2011, Google unveiled Dart (neé "Dash") at a conference.  There are numerous opinions about the "true" objectives of the Dart effort by Google and the author will not comment on these as they are largely speculation or self-serving commentary or both.  More definitively, Dart has some optional typing that gives one at least some of the benefits of a statically-type language.  The fact that these are optional and not enforced until run-time is far from the choice the author would have made.  However, given the current state of widely-applicable, in-browser technology, Dart is easily the least worst option.  (The only other option even considered for Seven5 was using Go on the client-side via the "PNaCl" interfaces, and that choice would have prevented any hope of running in multiple browsers.) Seven5 would continue to search for other alternatives were it not for the excellent work done by the Dart team in compiling Dart to cross-browser and efficient Javascript.

In December 2005, the firm 37 Signals released Ruby on Rails, or just "Rails."  It has had numerous version since then, with a 4.0 version arriving in mid-2013. Rails has been a huge success by any metric one could use to measure an open-source project. Commercially, it has had a certainly positive, but difficult to measure, impact on the products of 37 Signals.  Rails should be credited, in the author's opinion, with the success of Ruby as the former drove the latter, not the reverse, as one might expect and has in fact been observed with similar tools such as Django and the Python language. Rails must be credited for popularizing, not inventing, three key ideas that have been blatantly stolen by Seven5: 

* Don't repeat yourself (written as "DRY")
* Convention not configuration
* Making coding easy for the developer and the rest will follow

Some readers may see the third as different from the first two and all three of these will be discussed, and the
inclusion of the last hopefully justified, later.

## Reactionary revolutions?

In politics, one typically associates "reactionary" with elements of a power structure who seek to prevent change.  Thus, the combination "reactionary revolution" would be silly and "reactionary counter-revolution" would not. However, in all three of the cases above the authors and supporters of the tools mentioned seek to create a revolution--or perhaps less grandly, a "sea-change" in the minds of developers--and are reacting to some existing technology.  

Go is reacting to C++.  It takes little "reading between the lines" to see that the Go team was unhappy with using C++ and felt it was harming the understanding, or perhaps even "image", of statically-typed languages more broadly.  Rob Pike's blog post, for example, about the origins of Go presents a cleaned-up-for-public-consumption version of the story of designing Go based on what was wrong with C++ and C.  Since Go became public, Pike and others present Java as part of the "understanding of static languages" problem, although this inclusion will not bear close scrutiny and seems certain to be a way to appear less dogmatic in their reaction to C++.  Rob Pike, as one of the originators of C and implementor of the first C compiler, may feel that C++ has harmed his legacy.

Dart is reacting to Javascript.  This has been explicit since the beginning of the project.  The primary designers/implementors of Dart have come from the team that implemented V8, the exceedingly fast javascript compiler that runs inside Chrome and made node.js possible.  They clearly stated the issues that they felt were preventing Javascript from being sufficient for their, and others', needs and it cannot be argued that they don't sufficiently understand Javascript to know where the problems are.

Rails is reacting to Oracle's (neé Sun's) Enterprise Java Beans (EJBs) and related technologies.  Certainly in the early versions of Rails, policy choices such as "convention not configuration" were the opposite of the "configuration at any cost" mentality of EJBs that was then popular.  It is ironic that now Rails may  have become what it beheld:  _it_ has spawned reactions of so-called "micro-frameworks" such as Camping in the Ruby community; similarly with Django and Flask in the Python community.  The author feels that the without Rails' and its legion of  imitators' reaction to EJBs, the web would be a far less interesting place as many otherwise productive developers would abandon web development in disgust with being forced to use EJBs and the related cruft.

The author apologizes to those involved in the history discussed above; much has been elided in the interest of space and it is certainly an oversimplification to claim that projects of these sizes are motivated by simply a reaction to a solitary thing. With that apology in mind, seven5 has been created in support for the first two things above and as a reaction to modern web development problems that is similar in spirit, but different in practice, from Rails' reaction to EJBs.

## Strong-typing and development for the web

Seven5 supports and builds on Rob Pike's claim about the connection between dynamic typing, largely with interpreted languages, and fun.  In particular, it is hard to argue that the Web is largely the domain of dynamic typing.  With the notable exception of Java on the server-side, Web development runs on technologies like Rails/Ruby, Django/Python, and numerous others that use dynamism plus interpreters on the server. Given the lone exception of Java in this morass of dynamic types and Pike's suggestion, even if politically motivated, that Java is part of problem, _Seven5 seeks to make server-side web development in Go at least as fun as Rails/Django_.  In particular, the use of strong typing should prevent bugs, ease the testing burden, and generally allow more rapid development than Rails/Django.  If the reader is wondering how web development could be (somehow! gasp!) faster than Rails/Django, the case to consider is making major changes to the server-side code that involves nearly everything about how a website works.  With static typing, this is not a problem because most errors in the _rewhack_ will be caught by the compiler. It further reduces the perceived need, perhaps even fear, that is present with dynamic tools to "test everything three ways" so that major changes have a hope of being made correctly and without regressions.

Javascript is really the only choice on the client side of the wire, although there are some reasonably popular technologies that _target_ it rather than having the user program in it, notably Coffeescript and JWT.   Both of these tools were considered as client-side technology, but were deemed less attractive than Dart.   Numerous toolkits (backbone.js, agile.js, etc) have been built on top of Javascript to ease modern web programming tasks, further pushing the dynamism on the client-side although they do not change the fundamentals of Javascript. The author's objection to Javascript and support for Dart boils down to a single claim:

>>>>Javascript is counter-educational. The more one knows about programming languages or the more experience on has with multiple programming languages, the more there is to unlearn when developing in Javascript.

The author will not bore the reader with the myriad of deceptive (to the well-educated) elements in Javascript and simply observe that the existence of a 176-page book called _Javascript: The Good Parts_, that attempts to instruct one on how to avoid the bad parts of Javascript, seems to have been unnecessary for other programming languages with the exception C++ ( _Effective C++_ and sequels, 3 volumes, several editions, 1000+ pages).  Seven5's support for Dart is fundamentally a pragmatic decision based on the picking the strongest type-system that is available in all modern browsers.

It should be noted that at the time of this writing, _Seven5_ works on the client side with various emerging web standards, such as model-driven views.  This is not _Seven5_ technology _per se_, but rather that _Seven5_ has an idiomatic and with any luck useful way of using these technologies.  All of these technologies are "moving targets" so the reader would be well served to be careful about the particular Dart and library versions that are expected at various points in the text.  The author strives to keep the [Setup chapter](setup.html) up to date in terms of Dart SDK version being used, and from this most of the client-side dependencies flow.

### Seven5 and Rails

Seven5's relationship to Rails is part thievery and part spiritual.  In the two cases already mentioned, DRY and convention over configuration, Seven5 stole the Rails ideas wholesale and just adapted them to the programming languages in question.  Rails must be given full credit here.  In the case of DRY, Seven5 includes the ability to _cross_ programming languages and not repeat yourself, something that to the author's knowledge Rails does not support.   

At a high level, Rails was codifying the hard-won experience of its primary developer, David H. Hansen or DHH, into code and allowed that knowledge about the "way web apps have to work" to be exploited by many other developers.  Following the spirit of Rails development meant agreeing with the *opinions* of DHH (c.f. opinionated software).  It turned out that DHH's opinions are now largely regarded as "facts."  Seven5 has taken on some of the challenges that have been observed since Rails' debut in 2005 and form the opinion-base, or spirit, for Seven5's existence.

## Modern web applications

Seven5 makes the following key claims about the state of web development today.  If the reader strongly disagrees with any of these claims, the author advises that you stop reading this document.

1. Applications must have a web API, a.k.a. "everything is a web service"
2. Users expect highly interactive web applications with the interaction implemented on client-side (no page reloads)
3. Users' computers are largely idle and will become more so in the future

The first two of these are large-scale trends in web application development and hardly need much elaboration here.  If you want proof, examine nearly any web-based application introduced in 2011 or 2012.  Try clicking on the inevitable API button and looking at the source of the web pages to see if it includes a reference to jquery.

The latter of these is more conjecture than the first two, although it is derived from Moore's Law which has been reasonably reliable for some time.  In the large, accepting this claim argues that the burden of computation, whenever feasible, should shift to the client side of the web application as there are significant and under-utilized  compute cycles available there.  In more concrete terms, a Seven5 application will never both rendering a graph on the server, the server will simply send a list of points to the client side for rendering.  (Given assumption 2, it suggests also that the user will expect the ability to zoom in and out on different parts of the graph.)

### Consequences

Taking all three of these assumptions together, what are the consequences?

>>>> Seven5 provides no mechanism for server-side generation of content (#3 and #2).  

>>>> Seven5 makes server-side web development overwhelmingly about implementing a REST API (#1).

The first of these is the most surprising to most developers.  There is no templating language for Seven5, despite the fact that there is one (two?) in Rails, one in Django, and legions in the various Java toolkits.  Naturally, the EJB standard has an entire substandard (JSTL), plus multiple implementation choices and variants, regarding server-side templating.  Hopefully, this makes more clear importance of the author's comment above about the "wait and see" attitude towards the web-ui project within Dart, there must be _some_ way to generate HTML content in a web application.

The second consequence is largely to simplify the problem.  Seven5 simply defines the web browser "app" as a client of the server-side API to insure that the first challenge is met.  Most Seven5-based applications will not have any code in their server other than the implementation of the various REST resources needed to power the client.  However, this approach does lead to a "programmable API" without any changes to the server, should new clients or functionality be desired.   Another simplification is that "only implement REST services" simplifies a number of the type-system issues around how Seven5 applications are implemented and tested and these will be discussed later.

The choice of REST here was not particularly a vote of support for REST as much as an attempt to use a simple, well-understood wire protocol rather than trying to invent a new one.  Seven5, internally, has hooks for implementations that want to use a different strategy for their resources, such as XML-RPC, base64 encoded Gobs, or the horror that is SOAP.

## Step 1: Define the wire type

The first step in building a Seven5 app is to define the wire type.  This is done in Go and the struct defined will be the type exchanged over the wire between the client- and server-side of the application.  If you prefer, this can also be titled "API design" since with REST resources, this is the only thing needed to have an API.

{% highlight console %}

type Greeting struct {
    Id      seven5.Id
	Phrase  seven5.String255
	Gesture seven5.Boolean
}
{% endhighlight %}

>>> All the types used in a wire type declaration _must_ come from the package seven5.  These
can be found in `seven5/types.go`.  This is enforced at encoding- and decoding-time, not at compile-time.

>>> Wire types must have a field called Id that is of type seven5.Id. This is 64 bit integer uniquely identifying the resource value in question.

>>> Wire types may have substructures.  Substructures must contain only types from the seven5 package, but do not need an Id field.

You can [click here](#step_2_define_a_resource) if you want to skip to the next section's code and ignore the rationale.

The first of the restrictions on wire types appears at first to be the most draconian.  This restriction is justified for two reasons.  First, the intent of the wire type is to define the REST resource's public api and the types in the seven5 set of types are designed to show intention not just representation.  REST has no self-description mechanism so, the wire types are basically "all you got."

Second, the wire type must be unambiguously convertible to and from json which is a substantially less powerful type system than either Dart or Go.  Thus, adding the intention in Go helps Seven5 correctly generate Dart types like a timestamp (just used for ordering) vs. a date-time (shown to user, must be GMT) even though both may be the same value in json.

The reason for the Id requirement is that Seven5 wants each value returned, deleted, or created to correl


### NOTE SELF : Main, choosing options, naming

### NOTE SELF: Project Layout

### NOTE SELF: Internal errors

### NOTE SELF: Lots of public methods and fields

### NOTE SELF: Dispatchers touch the wire, resources don't

### NOTE SELF: PBundles shouldn't use headers
