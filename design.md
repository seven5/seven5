
Seven 5
=======

**Seven5** is a web microframework for go. It has a number of design objectives:

* Don't repeat yourself.  Don't repeat others' work; remove an item from the  
stack if it exists elsewhere.
* Use go's strong typing and fast compiler as an advantage for web developers.
* Make all parts of the application easy to test by a single developer on a 
single machine.
* Make the resulting system deploy and scale nicely on cloud-like 
infrastructure.

Recently Updated
-------

[Why Mongrel2?](why_mongrel2.html)

[Project Layout?](project_layout.html)

Data Model
----------
**Seven5** has neither the inevitable ORM nor SQL mapping so often associated
with other frameworks.  99% of the world doesn't need SQL for a website.  Because 
of convenient ORMs, developers end up with SQL databases simply because that's 
the easiest path in the web framework.

**Seven5** makes a different strategy easy.  Your data model is just, well,
your data model.  You use go's data structures and store them, unmodified,
in a big blob of RAM called [memcached](http://memcached.org/).  There is no 
disk storage, because disk is slow, expensive to manage, and generally more 
trouble than it's worth in the age of server machines with 64 or 128GB of 
RAM.  **Seven5** allows easy provision of 
redundant memcacheds in different locales, if you are really paranoid about
server crash.  You can, of course, make periodic backups if you choose to 
(although don't be surprised if we don't recommend it).  Memcached is directly 
supported today by cloud infrastructures like Amazon EC2 for production systems.

By default, we use [gobs](http://blog.golang.org/2011/03/gobs-of-data.html) for
storing our data because [Rob Pike](http://en.wikipedia.org/wiki/Rob_Pike) is a
righteous dude. If you prefer to flatten your datastructures to bytes with 
things like [protobufs](http://en.wikipedia.org/wiki/Protobuf) or even 
[JSON](http://json.org) that's up to you.


Guises
------

A notion that is critical to **Seven5**'s operation is the notion of a 'guise' 
(rhymes with cheese, not fries).  A guise is a bit of code that allows a 
computation to look like a file, a part of a file or other http-level 
resource.  Some important 
guise types are detailed  below.  A correctly written guise insures that any 
input it needs from the filesystem is always "up to date."  Further, 
correctly-written guises cache their results in memory.   These two properties 
insure there is never any confusion for the developer of the form 
"which version of the file is this?"

A simple, fictional guise is the `WhatTimeIsItInParisGuise`.  This guise takes
requests for any gif file and returns an image of a clock with the hands set to
the appropriate current time in Paris. (It's easy to do with go's image file
support!)

Two DSLs And A Microphone
-------------------------

**Seven5** has three carefully interlocking pieces to make web app development
more pleasant, and blindingly fast. _Time may, in fact, slow down for you because
of the speed of your development._  The first two are 
[DSLs](http://en.wikipedia.org/wiki/Domain-specific_language), implemented
as go entities, that generate
CSS and HTML, respectively.  The third is a carefully crafted Javascript 
library that understands how to interact with the results of the **source**
of these two DSLs.  All of these are implemented as Guises; the final one, 
though, is visible to a developer only by programming in Javascript.


### CSSGuise

We will first discuss the `CSSGuise` and its accompanying DSL.  This guise takes
an input which is a CSS file to return, such as `foo.css`, from the browser. It
finds the appropriate CSS file in the current project and combines that text
(unmodified) with the output of the DSL.  Let's look at an example of the DSL, 
implemented in go code.

	var (
		On = BinaryClassSelector{attribute:"visible", trueValue:"yes", falseValue:"no"}
	)
	
	func Foo_css() {
		return ON.css()
	}
	
When the browser/client receives the CSS resulting from its request, it is the 
contents of the static file plus this suffix: 

	.On { visible: yes}
	._On { visible: no}

Some key things to note about this simple example: 
	
* Changes in the go code are immediately reflected in the CSS code. Once
and only once.  It's hard, but not impossible, to end up with a program that
uses the wrong string for "On" as a CSS class name.

* The DSL allows common idioms to be expressed cleanly.  In our example here,
we are using the common CSS approach of having two classes that control whether
a particular object is visible on the screen.  Because it's encoded as an idiom
(see below) there is no way to become confused... is the class name
for turning something off `invisible` or `notvisible` or `not_visible`?  The 
`BinaryChoiceSelector` creates the second choice, `_On` programmatically.

* The resulting code in CSS uses names that are identical to the names in the
go code (via go's reflection mechanism).  You can say "css class 'capital-O-n'
is all frobbed up" to another developer without ambiguity.

* Because CSS has no run-time component, the DSL output is controlled by a 
simple function call, css() on an object you want to use in your app.

* Objects such as `On` arrive in the go code because are needed 
[**by programmers**](http://programming-motherfucker.com/).  If other types
things are needed by web designers, graphic artists, or other parts of the
team, they should go in the static file, `foo.css`.


### HTML Guise

The DSL associated with the HTML guise takes the place of "templates" in other
web frameworks. Consider this template, in the go template language:

	{{if user_is_logged_in $some_context}} 
	<strong>Hello {{.User}}, welcome to my world!</strong>
	{{else}} 
	<strong>Hello Anonymous</strong>
	{{/* fixme: need to add some CSS */}}
	{{end}}

**Seven5** takes the position that this is broken and wrong. It is not lost on 
**Seven5** that many people have also identified the pain and suffering that
templating like the above causes, particularly as websites grow to reasonable
sizes or have several different developers.  The **Seven5** developers 
smile slightly in quixotic reflection at the efforts of 
[mustache](http://mustache.github.com/): an effort to build a templating 
system without the pain of a templating system. Tilt on, brothers!

The author of the template above is attempting to do programming, not write HTML 
code---note the 'fixme' comment to remind him or herself go back 
later!   **Seven5**'s DSL for HTML means that
code uses the best tools and practices we know of for building software.  Put
in the negative, "How can a development environment or a best practice like
once and only once help you with the conflation of ideas and technologies in the
cesspool above?"

Let's consider the same attempted programming task using the HTML guise's DSL:

	var (
		Welcome = NewId()  // create a unique ID within the page
		WelcomeSection = Div(Welcome, On) // On? See above!
	)
	
	func welcome_message(ctx Context) {
		if is_logged_in(ctx[User]) {
			return WelcomeSection.html(Strong("Hello," + ctx[User]+ "welcome to my world!"))
		} else {
			return WelcomeSection.html(Strong("Hello, Anonymous welcome to my world!"))
		}
	}

One can certainly make the argument that writing a "template" in this form 
keeps other parts of the project team who do not code in go "out in the cold."
This is both correct and proper.   **Seven5** states that the task being 
attempted above is programming, and should be dealt with as such; in the amount
of time you save by not messing with stupid template files, you can teach people
enough to write in the DSL!

### How Does It Work?

The DSL above is actually translated, by the `HTMLGuise`, into the horrific
template shown originally! The construction of the `HTMLGuise` is analogous to 
the poor sod who had to write the go compiler; his pain of dealing with a far
lower language--assembly--benefits a great many people if he gets it right
one time and encodes it in a tool.

The above strategy has two implications that may be somewhat startling for
those unaccustomed to this approach.  First, the project's *source code* must
be available at run-time so the `HTMLGuise` can do its job. The source code
typically is not visible to external users of the application, but it must be 
visible to **Seven5**.  Second, to do its job perfectly the `HTMLGuise` must 
be able to understand arbitrary go code--for example to do correct 
type-checking.  In practice, only a subset of go can processed by the 
`HTMLGuise` and it is possible that the HTMLGuise can fail to produce a 
valid result.  If you don't like this, help improve the HTMLGuise to 
understand the go features you want!

### JSGuise, or 'The Other Side Of The Wire'

>Nota bene: This section has been heavily influenced by the Javascript
>experience of TrevorFSmith and KenFishkin.  IanSmith doesn't really know
>anything about Javascript.

"Poignard" is a Javascript toolkit that is spiritually similar to tools like
[JQuery](http://jquery.org/) but designed carefully to work with **Seven5**
applications---and **Seven5**'s DSLs for CSS and HTML.  Let's try a simple 
example that would be coded by a Poignard developer:

	function toggleWelcome() {
		var on = GO_Welcome().hasCSSClass(GO_On(true))
		GO_Welcome().dropAllCSSClasses()
		
		if (on) {
			GO_Welcome().addCSSClass(GO_On(false)) 
		} else {
			GO_Welcome().addCSSClass(GO_On(true)) 
		}
	}

> Perhaps Poignard should be a layer on top of JQuery?

The above example shows how the Javascript layer can be hooked to go language
entities.  A function like `GO_Welcome()` has as its definition the necessary
Javascript code to select the "Welcome" node (a `div`, see example above!)
from the DOM of the page. 

For a Javascript file request, it is the responsibility of the `JSGuise` to
examine the DSLs for both the CSS and HTML used for the page that the 
Poignard code is attached to.  It then suffixes the file, say `welcome.js`, 
with the additional function definitions necessary to access the 
DSL-plus-go-defined items at run-time in a browser.


Testing Javascript Code In Go
-----------------------------

I'm not really sure how to make this work in practice.  Here's a couple
of examples of things I'd like to write, written in english rather than as go 
code:

> Set the contents of the username field to "".  Verify that the continue button
is disabled.

> Set the contents of the username field to "ian".  Verify that there is
> a drop down present. If there is a dropdown present, verify that it has 
> exactly one item in it, with the contents "iansmith."

My goal would be that unit tests could be written in go, referencing the objects 
used in the DSLs for CSS and HTML and somehow have it "drive" the Poignard 
code through its paces.  I have the sense that the right way to do this is to
have Poignard abstract the notion of JS events slightly and allow these to be 
synthesized by the test harness.  I definitely do not want some crap like Selenium or
other "browser level" test harnesses.  **Seven5** went to a lot of trouble for 
once and only once, and the tests should benefit as well.

Easy case: changing the go code in the DSL of CSS or HTML
should cause immediate problems in the source code of the tests.  This is easy 
because the go compiler can check this and your IDE will tell you about it 
right away.  Because of the `JSGuise`, the now changed entities should cause 
the JS code to fail horribly--but not until the first time you run it.

More difficult to see is how changes in the JS code can be reflected back to
the go language tests automatically.  I suppose there are two basic options:

* Go code runs the show.  Since the go code is running the tests, it must be
told about what/how to access `Poignard` functions, etc.  This leads to a
once and only once violation since it requires that some entities be duplicated
from the Javascript world to the go world.  This is not _uberbad_ because the
test code runs and checks that things are ok so if you make a mistake at least
it gets caught fast.

* Write some type of Javascript analysis tool and use that to make entites,
at least functions, visible to the go level. There are some grammars 
[laying around](http://www.antlr.org/grammar/1153976512034/ecmascriptA3.g)
that could be used to extract some things from Javascript source and make them
visible to test code.  

Along this latter line, but without the analysis, perhaps there could be some 
simple rules associated with Poignard code.  Roughly, "Poignard can only
respond to events" and these are handled according to _blah_ convention.  Then
the test code could send synthetic events from the server to the 
client to drive the JS code.  Similarly, JSGuise could output some "testing
functions" that could be called from test code (via a network message to the 
browser from the server).

AJAX Stuff
-----------

Mongrel2 [already has](https://gist.github.com/920729)
support for the WebSockets proposal by 
[HyBi working group]
(http://tools.ietf.org/html/draft-ietf-hybi-thewebsocketprotocol-07) as well
as support for flash-based socket communication with JSON, XML, and blah blah
blah.  We should get this for free via our mongrel2 connection.

**Seven5** needs to exploit the client/server separation carefully to allow
unit tests to drive both client and server.


Naming
------

The framework is called **Seven5** because the originator lives in Paris, France.
All the postal codes for Paris, proper, begin with 75.  Besides, names don't
matter that much.

The use of the strange pronounciation of guise is because it sounds cooler.
Plus, the originator lives very close to the residence (compound?) of the 
[House de Guise](http://en.wikipedia.org/wiki/House_of_Guise) which is 
pronounced in this way.   The English word that is spelled the same way comes
from the rumor that a dis_guise_ was used by the Duc an attempt to mask his
involvement in the attempted assassination of 
[Gaspard de Coligny](http://en.wikipedia.org/wiki/Gaspard_de_Coligny) that
lead directly to the 
[St. Bartholomew's Day Massacre]
(http://en.wikipedia.org/wiki/St._Bartholomew's_Day_massacre). Web frameworks 
may educate in many ways.

The use of the name 
[Poignard](http://en.wikipedia.org/wiki/Poignard) is because sharp, narrow tools 
are often needed when working with web frameworks.  Such tools, correctly 
applied, can be the killing stroke whereas huge, but relatively blunt, tools
like Javascript often provide less death.  More death is better.


