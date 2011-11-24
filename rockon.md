# Seven5: Rock On!

	You may, at your sole discretion, make the Ozzy hand gesture while reading this document
	if you strongly agree with its claims or approve of its features.  If you choose to,
	you may also raise the Ozzy hand gesture above your head while performing the "synchronized
	head banging maneuver," although this is not recommended for the slower students.
		
	If you really believe that a feature of rock_on is truly world-changing or a near-death 
	experience, you may also attempt the rarely-seen "doubly Ozzy."  In this gesture 
	you make the Ozzy hand gesture with both hands simultaneously, with palms forward 
	or backward (respectively), while waving both above your head--synchronously.  
	
	Keep Rock Evil.
	

`rock_on` is the entry point to the magic of *Seven5*.  `rock_on` is responsible for doing all the necessary work to keep your project up-to-date with *Seven5* love; it provides all the _extras_ that make working with *Seven5* fun, as well as interacting with various external services.  Some services need a little "poke" to handle the speed of *Seven5* development.

>Trevor: The next part you may not agree with. Please advise.
	
Before you can "rock on!" you must run `gb -t` at the root of your project's directory structure.  This will build your project into a go package (`.a` file) and run all tests.  There's no point in running `rock_on` if you don't have a project that builds and passes all tests as these are problems you have to fix in any case.  We could have `rock_on` compile your project and run tests, but that would be repeating the work of fine go-based build tools like `gb`.  If you are the type that wants "one command to be all that and a bag of chips," try `gb -t && rock_on`.  (The test infrastructure of *Seven5* is carefully crafted to not slow down your development cycle; you *can* afford to run all tests on every build, trust us.)

`rock_on` has four main areas of responsibility:

1.  Validate the project's structure ("does it meet our standard project layout?") and discover all the entities you are using to which the *Seven5* conventions apply.
2.  Generate the necessary go code--based on what was discovered in the previous step.  Compile and link this generated code with your package created above into a runnable program.
3.  If requested, synchronize resources that are "in the network" (not on your local dev machine).  This step is optional since you may not have network connectivity when doing your development, or the network resources may be networkologically distant, and thus interacting with them could slow down the break-neck speed of your coding.
4.  Run the executable generated in step 2.  

With `rock_on` running on a modern laptop, you should expect `rock_on` to take less than 1 second. Yes, o-n-e second. Once you have proclaimed "rock on!", you can reload any web pages in your browser to see your changes. 

Thus, the development cycle for a *Seven5* developer is this:

* Make changes to your source go code using your favorite programmer's editor.

* `gb -t` to run all tests and make sure you didn't break anything.

* `rock_on  [project directory]` to bring everything up to date.

* Reload your web pages.

* Crack open a beer and enjoy.  Repeat as necessary.

An Example
==========

Discovery
---------

In step 1 above, `rock_on` scans your project's files looking for conventions.  For this example, we will assume that you have a file called `foo_rawhttp.go` and a file called `bar.css.go` in the root of your project directory.  Both of these are conventions of *Seven5* that `rock_on` knows about, so it swings into action.

Code Generation
---------------

Given what it discovered, `rock_on` must do two code generation actions: First, it needs to generate an executable program that "binds" your raw http-level handler to mongrel2 so it handles requests destined for `/foo/whatever`.  Second, it needs to generate javascript code based on the DSL inside the file `bar.js.go`.

In the first case, `rock_on` generates a `main()` function that hooks a type called `Foo` (note capitalization convention) to mongrel2 as an http handler.  Mongrel2 has a special protocol for such handlers, and `Foo` is expected to understand this protocol.  The resulting main is compiled and held for step 4, coming right up!

The second code generation task is more complex.  *Seven5* wants to give you the impression that the web is 'just go' all the way down!  The reality, sadly, is that there are less palatable tools involved, but `rock_on` and *Seven5* hide them from you.  When doing development of something like `bar.css.go` you get the strong typing and "once and only once" of go, but the result is a `.css` file that a browser can consume.  When doing *Seven5* development, never write a `.css` file by hand: Either you are doing something wrong or the framework is broken.  Both are better remedied by hacking go code.

Since CSS has no conditionals, the code generation process is quite simple.  `rock_on` takes the go function `bar()` from within `bar.css.go`, creates a new temporary file with a `main()` that has the contents of the original `bar()`, compiles it, executes it, and the output of the program's standard output is the CSS file.  `rock_on` puts this result into the proper place in the project (the `static` directory) as if it was written by a human.

Network Sync
---------------

*Seven5* expects you to keep your static resources, such as the CSS generated previously, but also all your graphics and text in a Content Delivery Network (CDN) such as Amozon's S3.  If you pass the `-net` option to `rock_on`, it will check the last update times of all static resources in the CDN and update them as necessary.  

> Trevor: We will need to think what happens if the developer chooses to not do this and just wants to use the static directory for content.  This is a reasonable choice for many people but we need to generate different URLs.  Probably should be a global switch of some kind, perhaps the developer can change the name of `static` to `static.no_net` to indicate they never want CDN delivery.

>Trevor: This is also the phase of `rock_on` where I'm imagining that amazon machines could be either booted or updated. It would be _super_sweet_ if somehow we "marked" the amazon nodes (tiny sized and free by default, of course) in some way with a property that says "managed by rock_on" and had an open "control channel" to the node where we could do updates *super* fast--maybe just be shooting a tarball of the project through the channel.

Restart
-------

`rock_on` now runs the executable it generated in step 2 and your changes are visible via 
a web browser `localhost:6767`.










