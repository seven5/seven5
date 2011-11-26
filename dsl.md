### Aside About Type Discovery

Because the HTML or any other guise would need to know what resources to serve, we are back to the problem of type discovery.  The real solution is, of course, to machine generate a snippet of code that "glues together" the developer package (`.a` file) and seven5 packages (as `.a` files) with a `main.main()` function. I propose this temporary solution:  

* Use the `init()` method of the file to "hook" a file of developer code to the toolkit. `init()` is automatically run at load-time. 
* The toolkit calls should be super-simple and able to take multiple parameter types.

The latter is to ease
the machine-generation process later and avoid having to do fancy things like understand types in go for real.

This is an example from `sass-examples.css.go` currently in the `samples/dsl` project:

	func init() {
		//note: Any "var" at top level should get shoved into here... could probably even
		//note: generate this init method this with some horrible shell script with some
		//note: sed that looks for var and then =
		Reg(contentNavigation)
		Reg(border)
		Reg(heavyList)
		Reg(errorIntrusion)  
		Reg(error)
		Reg(badError) 
		Reg(dataTableBase)
		Reg(dataLeft)
		CSS("sass_examples")//signal that you have told it everything
	}

Proposal on DSLs
================

HTML
----

There are two versions of the HTML DSL idea, _cheap and easy_, and _expensive and hard_.  Naturally, the latter is preferable but will take a lot longer to build.  We'll just call them "simple" and "complex" in here.

For simple, I'm proposing that `[anything].html` outside of `static/` gets sent to the `HtmlGuise`.  This guise looks up his known list of pages and if he knows the page, he renders it. Otherwise he gives 404.  (I do think an `ErrorGuise` might somehow be a good idea to concentrate all the logic of that in one place, especially the ability to turn on or off debugging messages, etc.)

The `HtmlGuise`, in the simple version, just finds the object and calls `dumpMeYourDamnHtml()` on it. *There is no need for, and no utility to, generating this to disk.*  The "DSL part" is all about trying to make writing the implementation of that nicer.  For simple, I'm proposing something dead-stupid:

	page:=new(Page)
	page.add(new Head().add(new (StyleSheet).setName(MYSTYLESHEET)),
	         new Body().add(...)
	
I think you see how this works.  The only way I see to make this palatable is to permit a wide variety of argument "types" to be used in various places, plus any number of arguments, to make it flow more nicely.  Go has pretty good support for methods that take any number of arguments and a wide range types--using reflection to figure out what to do with the parameters as you go.

Probably something much better/nicer/pretty/more terse can be done that the crap I have above. For me the key thing is that something like a stylesheet name, or a div id is an actual go variable.  OAOO.  If the typing is right I'd like to be able to say:

	const NAVBAR = DomId("nav")  //DomId 'is' a string but typechecked as a different type
	
and then use that in both the creation of the HTML--where it gets slapped onto a Div object--and in the CSS where it gets other love (see below, it is already in place at the current time).

In this simple version, the "code" of the page has to be rendered (run) completely every time.  Consider this:

	if inbox.Empty() {
		page.add(...some message that nobody loves you and you have no mail...)
	} else {
		page.add(...table headers...)
		for i,m:=range(inbox.Messages()) {
			if i%2=0 {
				page.add(...row based on properties of m...gray background)
			} else {
				page.add(...row based on properties of m...white background)
			}
		page.add(...table footers...)
	}

Normally, this kind of cruft is done inside the templating mechanism of the web framework--which is why I rebel and say that this is really programming.  However, this simple version is a bit verbose because you end up having to write your HTML in a weird, somewhat indirect way.  This is, roughly, _the same experience_ as templating since templating allows you to write the HTML naturally, but the code is written in a weird and indirect way.  I'd argue this is superior to templating only in that it allows OAOO for many critical things and--if you do it right--HTML coding which is basically the same task that you would do in a `.html` file anyway, except checked by the compiler for correct nesting and so forth.

### Possible area for opinion

We could make the decision that HTML code with conditionals in it is broken and wrong, and thus
the HTML guise and the DSL would get simpler and could be run only once.  This seems to be just
moving the problem to the Javascript code which then would have to do various DOM manipulations
when the inbox is empty, deal with pagination, etc.  Maybe that's ok, and then the HTML
DSL could be reduced to something broadly similar to the CSS DSL...

CSS
---

CSS has no dynamics, so it only needs the simple version of the DSL.  I implemented this as
`css.go` in the `seven5/css` package.  I really wanted to see if the notation was acceptable;
not sure if it is super "clean" but it does have the nice properties for programmers about 
type-checking, OAOO, development tools that can do completion on css fields, etc.  

This is a somewhat reduced version of the `sass-examples.go` CSS that is in `samples/dsl`.  It is a direct translation of the examples from the [homepage of SASS](http://sass-lang.com).  It expects that you really understand composite literals in go.

	//this is not really recommended (use of .) but it makes the DSL cleaner
	import . "seven5/css"

	//init() method removed, it's above

	//OAOO...used below
	const blue = 0x3bbfce

	//OAOO if we wanted to use in other parts of the program it is now visible... note the type
	//prevents this from being used in bogus places
	const CONTENT_NAV = CSSClass("content_navigation")

	var contentNavigation = ClassStyle{
		CONTENT_NAV, 
		Style{ 
			//shorthand 
			BorderColor{0x3bbfce}, 
			//more explicit
			Color{Rgb: 0x2b9eabe},
			},
		nil, //no inheritance... this nil is annoying but it can be avoided, see next example
		}


	//if you name the fields you ARE using, you can omit others so this
	//can be used to avoid the "extra" nil
	var border = ClassStyle{ Class: CSSClass("border"), 
		Style: Style{ 
			Padding{Px: 16},
			Margin{All: Size{Px:16}},
			BorderColor{blue},
		},
	}

	//functions are legal just like in sass
	func tableBase(id DomId) IdStyle {
		twoPixOfPadding:=Padding{Px:2}//OAOO!
		return IdStyle{Name: id, Style: Style{
			TH{TextAlign{Center:true}, FontWeight{Bold:true}, twoPixOfPadding},
			TD{twoPixOfPadding},
		}}
	}
	//better type checking than sass! only a size can be passed to function!
	func left(id DomId, size Size)  IdStyle{
		return IdStyle{Name: DomId(id), Style: Style{Float{Left:true},MarginLeft(size) }}
	}

	//note strong typing means you MUST convert this to a DomId to be able to use it! sweet!
	var data = DomId("data")
	var dataTableBase = tableBase(data)
	var dataLeft = left(data,Size{Px:10})

Javascript
----------

I don't know enough about javascript to really do much with this.  The only rough idea I had
was to arrange the go types that you "program against" in "foo.js.go" to work out to be very close
textually to the "looser typed" Javascript entities:

	func bar(p1 type1) {
		Alert("you suck:",p1)
	}

becomes

	function bar(p1) {
		Alert("you suck:",p1)
	}
	
There would be a very stupid machine translator that would do the work of this transformation--but it would _assume_ that you had already passed the go compiler.	In other words, it will be mostly a text-to-text translation, not really a compiler.
	
The hope would be that you could just do enough typing on the go side to get you the things you want to do on the javascript side. Note that the go code in this design is never actually _run_ so you don't have to implement things in go, just do enough to get the types right. 

Then seven5 would supply a "bridge" library of Javascript code that would aid the translator in its effort to be stupid by providing a version of `Alert()` that called `window.alert()` and shoved the parameters passed to `Alert()` into a string so `window.alert()` wouldn't complain about number of arguments... you get the idea.  I would hope that most dom manipulation could be supported with a translator that was _uberstupid_.

It should be clear that you can write a valid go programs in `foo.js.go` which translates to something that doesn't make sense, even syntactically, in javascript.  The key is finding the sweet spot of stupidity, particularly for control structures (like `range` in go) and the ability to do some function definitions in go that "map" to the JS side correctly.  

Obviously, if we are going to use `backbone.js` as our story for REST we need to provide go classes that "look like" backbone's objects, functions, etc.

The only actually "good thing" I thought of along this line is providing a "go side wrapper" (if you see my meaning) around jquery--plus the necessary support on the JS side--so that people that already know jquery will feel at home.  I just hate the jquery notation soooo much.

The Complex Version
-------------------

This idea applies mostly to the HTML DSL, although the level of translator sophistication needed to do this would probably help the Javascript DSL as well--if you go with something like the approach outlined above.  

In this approach, you actually need to do some analysis of the go code in the HTML DSL.  The objective is to try to find "snippets" (roughly [basic blocks](http://en.wikipedia.org/wiki/Basic_block)) that can be run to completion and the result cached.  Then a new version of the DSL code would get emitted that contained _only_ the necessary logic (control structures and such) and data that could possibly change and affect the outcome of the HTML generation.  So, a very complex page is reduced to only needing to compute (say) the inner loop of some varying data at run-time.  All the rest comes straight from cache!

Doing this right also means having a few conventions about how aggressive you want the cache to be.  For example:

	//This is the critical part of the home page! It's hit 10 times per second on average!
	for i_30secs, v:=range values {
		//generate wicked complex HTML here and compute 400 mersienne primes...
	}
	
With this in place, you can also debug in "always up to date mode" but when "going big" you know that each evaluation of the loop serves all page views for 30 seconds or something like that...

Need For The Complex Version
----------------------------


 






