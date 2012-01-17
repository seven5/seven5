# Seven5: Develop

<nav>
  <ul>
    <li>[Intro](index.html)</li>
    <li>[Install](install.html)</li>
    <li>[Develop](develop.html)</li>
    <li>[Pontificate](pontificate.html)</li>
  </ul>
</nav>

The instructions below assume that you have [already installed](install.html) all the necessary code and programs.  You need to be sure that all the environment variables such as GOROOT and GOPATH are set in the shell you are working in.

## Install Seven5 sample program: Dungheap

	cd $GOPATH/src
	git clone -b sample-dungheap git@github.com:seven5/seven5.git dungheap

The above command checks out the 'sample-dungheap' branch into the directory named dungheap. The resulting directory needs to be named exactly dungheap because the name of the directory and the package declaration in the Go files must agree.

## Notes on project directory/file structure

    dungheap
        |
        +-- link.bbone.go 
        |
        +-- pwd.go 
        |
        +-- dungheap_test.sqlite 
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
             |    +-- control (a unix domain socket)
             |
             +-- static
                  |
                  +-- index.html
                  |
                  +-- SpecRunner.html
                  |
                  +-- js
                  |    |
                  |    +-- dungheap.js
                  |    |
                  |    +-- dungheap-test.js
                  |
                  +-- vendor
                  |
                  +-- image
                  |
                  +-- css
                  |
                  +-- template

* Your project name should be the same as your package name in go.  The result of building your code should be a library (.a file).

* `link.bbone.go` the standard name for a type `Link` that is going to be exported to the client side as a backbone model ("bbone" for short).  This file contains the definition of the model as a Go struct and an implementation of the interface `seven5.Restful` that handles storage and retrieval of `Link` structs.  It is expected that this Restful service can be accessed with the method `dungheap.NewLinkSvc()`, a.k.a.  `projectName.New`ModelName`Svc()`.  

* `pwd.go` is a file that normally should not be checked into the repository.  This file contains private bootstrap information for the project such as creating the initial super-user for a web application.  It has been included here so developers can see what should be placed in this file.

* `dungheap_test.sqlite` contains configuration data that is needed for mongrel2 to connect to the application in test mode.  If you have the `sqlite3` application, you may examine the project configuration by looking at the tables inside this file.

* `mongrel2/log` contains logs (duh) including access, error, and logging from the app code.  Application logging goes to `seven5.log` and a logger connected to that file can be accessed with the `AppStarter` interface in go.

* `mongrel2/run` contains control files generated and used by mongrel2 to start, stop, and control the server. The `control` file is a unix-domain socket that is used to tell mongrel when we have changed the configuration or restarted your application.

* `static` contains files which will be served by mongrel2 unchanged. The web server will be chrooted to this directory.   Mongrel2 is smart enough to watch these files for changes and make sure it only serves up the latest versions and issues appropriate 304 results when the files have not changed.

* `js` contains the application javascript files.  In the case of dungheap, there is the application code and its accompanying tests.

* `SpecRunner.html` is a web page that you can request (http://localhost:6767/static/SpecRunner.html) to run the dungheap tests.

* `index.html` is the web page that starts up the dungheap application (http://localhost:6767/static/index.html)

* `vendor` has a number of carefully chosen javascript libraries for use with a *Seven5* application.  These are intended to be used primarily as the basis for the MVC support on the client side (via [Backbone](http://documentcloud.github.com/backbone/)) or to provide testing support (via [jasmine](http://pivotal.github.com/jasmine/)).

* `css` and `image` hold only static files for the application, style sheets and images respectively.

* `template` is part of the MVC structure of the dungheap application.  This directory holds snippets of HTML code that the client side can load (via ajax) to create various parts of the presentation.  These files are _static_ because they contain only data that is consumed by the client side.  Without this, large collections of HTML require tedious and error prone javascript to create solely on the client-side.

# Seven5: Rock On!

	You may, at your sole discretion, make the Ozzy hand gesture while reading this section
	if you strongly agree with its claims or approve of its features.  If you choose to,
	you may also raise the Ozzy hand gesture above your head while performing the "synchronized
	head banging maneuver," although this is not recommended for the slower students.
		
	If you really believe that a feature of `rock_on` is truly world-changing or a near-death 
	experience, you may also attempt the rarely-seen "doubly Ozzy."  In this gesture 
	you make the Ozzy hand gesture with both hands simultaneously, with palms forward 
	or backward (respectively), while waving both above your head--synchronously.  
	
	Keep Rock Evil.
	
`rock_on` is responsible for doing all the necessary work to keep your project up-to-date with *Seven5* love; it provides all the _extras_ that make working with *Seven5* fun, as well as interacting with various external services.  Some services need a little "poke" to handle the speed of *Seven5* development.

## Run the server

	cd $GOPATH/dungheap
	rock_on

Point your browser at [the dungheap application](http://127.0.0.1:6767/index.html) and you should see some URLs plus discussion about them.

## Make a change, feel the love

The `rock_on` command is a tool used during development to make the development cycle of Seven5 apps as fast as rum on a Kauai beach. It runs mongrel2, manages the seven5 server and recompiles your code anytime it changes (via the 'go' command).  `rock_on` uses the command `tune`  command (both are in your $GOPATH/bin directory) to generate a small main function that glues together your library (.a file) with the seven5 library (.a file).  Tune generates its code into the _seven5 directory; you can see the source code for the main program in there as _seven5/dungheap.go.

To see this in action, open your programmers' editor and point it at `link.bbone.go.` Add a space or make some change to the program, such as in the `Link` struct, and save the file.  On your command line you should see some messages from `rock_on` about recompiling and restarting. If you screw up the code and it won't compile, you'll see a helpful error message. Fix the code and save and then `rock_on` will do it again, and you can keep on rocking.

That's your dev cycle: save file, reload in the browser.  On any half-decent laptop, you should be able to get a rebuild from `rock_on` in about 2 seconds; on a nice laptop with an SSD, it's sub-1 second.


## Walkthrough of Link and the LinkSvc in Dungheap

TBD.

## Seven5 doesn't do windows

The *Seven5* server does no server side presentation handling.  Read the previous sentence again, because it means what it says and that's freaky for most web developers.  Literally, *Seven5* will not render a template or change HTML in any way as it goes over the wire.  It will serve up some blobs of html if it is static.

>You don't need it. Really.

*Seven5* serves up awesome APIs and automatically generates the Javascript to manipulate them.  All that's left is for you to write the UI in the native languages of the web: HTML, Javascript, and CSS.
