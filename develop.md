# Seven5: Develop

<nav>
  <ul>
    <li>[Intro](index.html)</li>
    <li>[Install](install.html)</li>
    <li>[Develop](develop.html)</li>
    <li>[Pontificate](pontificate.html)</li>
  </ul>
</nav>

> Editors note: A lot of this is speculative as we solidify the lingo and patterns.

The following steps will create a basic blog application named "blargh":

## Create the project directory:

  seven5-create-project blargh

This will create the following directory structure:

    blargh
        |
        +-- handlers.go 
        |
        +-- resources.go 
        |
        +-- sass
        |    |
        |    +-- site.sass 
        |
        +-- js
        |    |
        |    +-- site.js 
        |
        +-- mongrel2.sqlite 
        |
        +-- mongrel2
             |
             +-- log
             |    |
             |    +-- access.log 
             |    |
             |    +-- error.log 
             |    |
             |    +-- seven5.log 
             |     
             +-- run
             |    |
             |    +-- mongrel2.pid 
             |    |
             |    +-- run_control (a unix domain socket)
             |
             +-- static
                  |
                  +-- favicon.png


* `handlers.go`, `resources.go` and contain (respectively) the views and data. 

* `mongrel2.sqlite` contains configuration data that is needed for mongrel2 to connect to the application in test mode.

* `log` contains logs (duh) including access, error, and logging from the app code.

* `run` contains control files generated and used by the front end.

* `static` contains files which will be served by mongrel2 unchanged. The web server will be chrooted to this directory.

## Run the server

	cd blargh
	seven5dev

Point your browser at http://127.0.0.1:8000/ and you should see a nice Seven5 welcome page.

## Make a change, feel the love

The seven5dev command is a tool used during development to make the development cycle of Seven5 apps as fast as rum on a Kauai beach. It runs mongrel2, manages the seven5 server and recompiles your code when there are changes.

To see this in action, open your programmers' editor and point it at handlers.go. On line X, change the name of the RootHandler to BlarghHandler and then save the file. On your command line you should see some messages from seven5dev about recompiling. If you screw up the code and it won't compile, you'll see a helpful error message. Fix the code and save and then seven5dev will fix it up and keep trucking.

That's your dev cycle: save file, reload.

Now point your browser at http://127.0.0.1:8000/blargh/ and note that Seven5 is dynamically routing URLs to handlers based on their names. Aw, yeah!

## Create the blargh and blabs

In resources.go create structs like so:

	type Blargh struct {
		Title, Author string
		...
	}
	type Blab struct {
		Title, Name string
		...
	}

> Describe how to CRUD and query .  Slam ORMs some more.


## Two DSLs and a microphone

Seven5 has three carefully interlocking pieces to make web app development more pleasant and blindingly fast.

_Time may, in fact, slow down for you because of the speed of your development._

The first two are [domain specific languages](http://en.wikipedia.org/wiki/Domain-specific_language) (DSLs) for HTML and CSS. They're implemented as go entities, are compiled to native code and they quicken the development of valid pages. The third piece is a carefully crafted Javascript library, Poignard, that understands how to interact with your go data and the compiled results of the DSLs.

Altogether, these three pieces automate much of the boilerplate work required by less opinionated frameworks.

### Guises

Both the DSLs and Poignard are implemented using a notion that is critical to Seven5's operation: the 'guise'.

The proper pronunciation of "guise" is like "geezer" without the "er".

> We need an audio clip of Ian [saying](http://www.paul.sladen.org/pronunciation/) "Hello, this is Ian Smith and I pronounce guise as guise"

A guise is a bit of go code that make computation look like a web resource in the [RESTful](http://tomayko.com/writings/rest-to-my-wife) sense.  The guises you'll use most are the HTML guise, the CSS guise, and the JS guise.  They implement all of the machinations of the DSLs and communication with Poignard.  There are other guises like the Auth guise which handles the dirty truth about who you are to Seven5, too.

## Enjoy the free HTML, CSS, Ajax and events

Anything which can be automated should be.  Yes, you can reach a flow state implementing yet another RESTful API and fronting it with JS objects and staying in sync via websockets.  You can also use chainsaws to carve bears statues out of stumps.  Pick a better hobby and don't do it on company time!

### HTML Guise

> Make these examples use the context of the blargh and set up the next step when we use Poignard

The DSL associated with the HTML guise takes the place of "templates" in other web frameworks. Consider this template, in the go template language:

	{{if user_is_logged_in $some_context}} 
		<strong>Hello {{.User}}, welcome to my world!</strong>
	{{else}} 
		<strong>Hello Anonymous</strong>
		{{/* fixme: need to add some CSS */}}
	{{end}}

Seven5 takes the position that this is broken and wrong. It is not lost on Seven5 that many people have also identified the pain and suffering that templating causes, particularly as websites grow to reasonable sizes or have several different developers. The Seven5 developers smile slightly in quixotic reflection at the efforts of [mustache](http://mustache.github.com/): an effort to build a templating system without the pain of a templating system. Tilt on, brothers!

The author of the template above is attempting to do programming, not write HTML code &mdash; note the 'fixme' comment to remind him or herself go back later!  Seven5's DSL for HTML means that code uses the best tools and practices we know of for building software. Put in the negative, "How can a development environment or a best practice like once and only once help you with the conflation of ideas and technologies in the cesspool above?"

Let's consider the same attempted programming task using the HTML guise's DSL:

	var (
		Welcome = NewId() // create a unique ID within the page
		WelcomeSection = Div(Welcome, On) // On? See above!
	)
	
	func welcome_message(ctx Context) {
		if is_logged_in(ctx[User]) {
			return WelcomeSection.html(Strong("Hello," + ctx[User]+ "welcome to my world!"))
		} else {
			return WelcomeSection.html(Strong("Hello, Anonymous welcome to my world!"))
		}
	}

One can certainly make the argument that writing a "template" in this form keeps other parts of the project team who do not code in go "out in the cold." This is both correct and proper.  Seven5 states that the task being attempted above is programming, and should be dealt with as such; in the amount of time you save by not messing with stupid template files, you can teach people enough to write in the DSL!

> mention the huge library of built in markup and style elements


### JSGuise

When writing a web app of any size there is always some moment where people decide to unify all of the organically grown client side data wrappers, events, and API calls.  It's painful and it's needless reengineering.  Seven5 provides all of this out of the box and does it with style.

First, let's add our blargh and blabs structs to the JS guise.  In resources.go:

> something like JSGuise.mapResource(Blab, ...)

"Poignard" is a the Javascript side of the house and is spiritually similar to tools like [Backbone.js](http://documentcloud.github.com/backbone/) but designed to interact with Seven5's HTML, CSS, and JS guises.

Let's try a simple example that would be coded by a Poignard developer:

> Make this example query for blabs and display it in the HTML we created above.

	function toggleWelcome() {
		var on = GO_Welcome().hasCSSClass(GO_On(true))
		GO_Welcome().dropAllCSSClasses()
		
		if (on) {
			GO_Welcome().addCSSClass(GO_On(false)) 
		} else {
			GO_Welcome().addCSSClass(GO_On(true)) 
		}
	}

The above example shows how the Javascript layer can be hooked to go language entities. A function like `GO_Welcome()` has as its definition the necessary Javascript code to select the "Welcome" node (a `div`, see example above!) from the DOM of the page. 

For a Javascript file request, it is the responsibility of the `JSGuise` to examine the DSLs for both the CSS and HTML used for the page that the Poignard code is attached to. It then suffixes the file, say `welcome.js`, with the additional function definitions necessary to access the DSL-plus-go-defined items at run-time in a browser.

### CSSGuise

> Make this mess with display of the Blabs we started in the preceding section

This guise takes an input which is a CSS file to return, such as `foo.css`, from the browser. It finds the appropriate CSS file in the current project and combines that text (unmodified) with the output of the DSL. Let's look at an example of the DSL, implemented in go code.

	var (
		On = BinaryClassSelector{attribute:"visible", trueValue:"yes", falseValue:"no"}
	)
	
	func Foo_css() {
		return ON.css()
	}
	
When the browser/client receives the CSS resulting from its request, it is the contents of the static file plus this suffix: 

	.On { visible: yes}
	._On { visible: no}

Some key things to note about this simple example: 
	
* Changes in the go code are immediately reflected in the CSS code. Once and only once. It's hard, but not impossible, to end up with a program that uses the wrong string for "On" as a CSS class name.

* The DSL allows common idioms to be expressed cleanly. In our example here, we are using the common CSS approach of having two classes that control whether a particular object is visible on the screen. Because it's encoded as an idiom (see below) there is no way to become confused... is the class name for turning something off `invisible` or `notvisible` or `not_visible`? The `BinaryChoiceSelector` creates the second choice, `_On` programmatically.

* The resulting code in CSS uses names that are identical to the names in the go code (via go's reflection mechanism). You can say "css class 'capital-O-n' is all frobbed up" to another developer without ambiguity.

* Because CSS has no run-time component, the DSL output is controlled by a simple function call, css() on an object you want to use in your app.

* Objects such as `On` arrive in the go code because are needed [**by programmers**](http://programming-motherfucker.com/). If other types things are needed by web designers, graphic artists, or other parts of the team, they should go in the static file, `foo.css`.

## Squirt it to the cloud

> TBD
