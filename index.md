
Seven 5
=======

Seven5 is a web microframework for go. It has a number of design objectives:

* Don't repeat yourself.  Don't repeat others work; remove an item from the one stack if it exists elsewhere.
* Use go's strong typing and fast compiler as an advantage for web developers.
* Make all parts of the application easy to test by a single developer on a single machine.
* Don't bother with parts of the stack that are needed in today's web development.

Guises
------

A notion that is critical to Seven5's operation is the notion of a 'guise' (pronounced "geez" not 
"guys").  A guise is a bit of code that allows a computation to look like a file or other http-level
resource.  Some example guise types are detailed below.  A correctly written guise insures that 
any input it needs from the filesystem is always "up to date."  Further, correctly-written guises 
cache their results in memory.   These two properties insure there is never any confusion for the
developer of the form "which version of the file is this?"

### CSS Guise

A CSS guise called the `SassGuise` allows HTTP-level resources like foo.css to be generated automatically from
a file like `foo.sass`.  

### Image Guises

An example guise has been provided, the `ParisTimeGuise` which will respond to requests for any gif
file (ignoring the requested name!) with an image of the current time in Paris.  This guise uses
the go library `svg` to render the image dynamically.  One can easily imagine a guise such as a
`BarCodeGuise` that generates a particular bar code, on the fly, needed for a web application or perhaps
from a csv file present on the filesystem.

### JS Guise

A critical guise that is supported by Seven5 is the `JavascriptGuise` for go.  This allows a developer
to do some simple programming tasks for the _client_ side of an application in go.  The code written
is compiled by the `JavascriptGuise` to javascript code when it is requested by a client.  The intent
of this guise is not to allow arbitrary Javascript code to be written, but as a simple "bridge" that
allows typical event handlers and DOM manipulation to be done in go.

If you are wondering how this compiler actually works, read on about vues and you'll figure it out.


Vues
----

Vues are not the same as "views" in many other web toolkits, although they perform a similar action.
The use of the spelling v-u-e is to emphasize the differences as well as suggest that a vue (unlike
a view?) takes a point of, well, view.  _Vue_ in French, like View in English, represents either 
something seen with the eyes **or** an opinion, "In my view, the Seahawks will win the SuperBowl
this year."  It seems that many web toolkits have forgotten the latter choice of meaning.

Vues are a programmatic way to express the desired output of an HTTP request.  They are related
to, but the inverse of, "templates" in many web development systems.  Consider this template (in
the syntax of go's template language):

	{{if user_is_logged_in $some_context}} 
	<strong>Hello {{.User}}, welcome to my world!</strong>
	{{else}} 
	<strong>Hello Anonymous,</string> would you like to <A HREF="newuser">create an account</A>
	{{/* fixme: need to add some CSS */}}
	{{end}}

The view (!) of Seven5 is that this is horrifying.  This opinion is shared by many other efforts,
such as the quixotic [mustache](http://mustache.github.com/), in their attempts to 'simplify', 'clarify', 
or 'improve' the template
writing experience.   Anyone who has written ASP, JSP, go templates, ERB, etc.  will 
certainly agree that at a minimum such work is unpleasant and at worst impossible.  The key difference
between all the other systems and Seven5 is this:

>>> Seven5 considers this a task for a software developer. 
 
With this difference clearly in mind,
it is natural to proceed to the idea that programming tools---unable to make progress with 
horrors like the template above---should be brought to bear.

In Seven5, deciding what result, typically HTML, should be emitted for a given request is the
business of a programmer.  This developer can, at his or her discretion, chose to do this in a
way that allows other folks--such as pixel weasels, web standards weasels, mobile browsing weasels--
to work closely with him or her.  To do so would almost certainly require well-structured 
HTML combined with carefully considered use of CSS. This type of output is supported, but not 
required, by Seven5. 

Consider the inevitable 'hello world example' of how to write a vue:

	package sample
	import "seven5"
	type HWVue struct {
		seven5.SimpleVue
	}
	func (self *HWVue) render(context map[string]interface{}) {
		self.Literal("hello world")
	}
	

(The above is a valid Seven5 application!)  However, because this is actually go code--not 
some horrible conflation of unstructured text and programming constructs--this is another 
vue that produces the same output.

	package sample
	import "seven5"
	type HWVue2 struct {
		seven5.SimpleVue
	}
	func (self *HWVue2) render(context map[string]interface{}) {
		self.Literal("hello wor"+string('l')+"d")
	}

Naturally, the weasels mentioned above will complain that they can't participate in the "design
process" because this "all in go code" and "we can't run the compiler" (why?).  Seven5 allows external 
resources, such as those written by said weasels, to be integrated into the backseat of the work 
of responding to a request; go code stays in the 
front seat.  Here's an example that may help you see how we might integrate other people's work:

	package sample
	import "seven5"
	type HWVue3 struct {
		seven5.SimpleVue
	}
	func (self *HWVue3) render(context map[string]interface{}) {

		self.Fragment("page_header.html")
		self.Markdown("long_bloated_intro.md")
	
		//add some CSS machinery to the literal result
		self.Literal("hello world",".helloclass","#helloid")
	
		self.Fragment("page_footer.html")
	}



### Times are a changing

A Vue expresses the *compile-time* information that is known about the result--usually some type
of layout or "framing" of the result--plus some logic about how to combine that information with information that
is known only at *run-time*, such as the current user of the website's name.  (You're a go 
programmer, you can handle this.)   Let's give an example to show the difference. 

	package sample
	import "seven5"
	type HWVue4 struct {
		seven5.SimpleVue
	}
	func (self *HWVue4) render(context map[string]interface{}) {
		self.Literal("hello world...")
	
		if context["user"]!=nil {
			self.Literal("...and welcome back!")
		}
	}

Given this vue, the text "hello world..." will be displayed no matter what.  This is known at "compile
time."  However, based on who the actual web requested is made by, some other HTML might be emitted that
welcomes the user back.  

How Does It Work?
-----------------

A vue is actually compiled by Seven5 into a go template (see above! ugh!) because, in fact, a go
template expresses exactly the same things---but in a way that doesn't let us bring our development
knowlege and tools to bear so easily.  The data needed by the vue is
fed to the resulting go template at the time the final result is needed, a.k.a. run-time.  

This "trick" is possible primarily because go is easy to parse and go comes with the 
necessary tools to do this "out of the box."  The vue "compiler" can also exploit the use of
eclipse and other IDEs that do automatic compilation and thus type-checking.  Because your
vue is part of go program that you are writing in eclipse, the type system applies to it,
allowing you to "check" your vue in a way you could never do with a template! 

The vue "compiler" can, and will, generate bogus templates that will die horribly *if* you fail to
make sure your go program, vues and all, compiles.  You _can_ do this with make, but you are
probably using something already that shows a little red stop sign whenever you type any
character that causes the compile to break.  So, the go functions in the examples above are actually
no-ops in terms of the "program" you are working on.  They are simply there to force typechecking,
correct numbers of arguments are present, etc.  The vue compiler will dutifully copy __whatever__
is inside the parenthesis of `Literal` without the least concern because it assumes that you
have run the standard compiler _a priori_.

