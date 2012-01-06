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
                  +-- index.html
                  |
                  +-- index.js

* `handlers.go`, `resources.go` and contain (respectively) the views and data. 

* `mongrel2.sqlite` contains configuration data that is needed for mongrel2 to connect to the application in test mode.

* `log` contains logs (duh) including access, error, and logging from the app code.

* `run` contains control files generated and used by the front end.

* `static` contains files which will be served by mongrel2 unchanged. The web server will be chrooted to this directory.

## Run the server

	cd blargh
	rock_on

Point your browser at http://127.0.0.1:8000/ and you should see a welcome page.

## Make a change, feel the love

The `rock_on` command is a tool used during development to make the development cycle of Seven5 apps as fast as rum on a Kauai beach. It runs mongrel2, manages the seven5 server and uses the `tune` command to recompiles your code when there are changes.

To see this in action, open your programmers' editor and point it at site.go. Change the Name field in the Site struct and save the file.  On your command line you should see some messages from `rock_on` about recompiling. If you screw up the code and it won't compile, you'll see a helpful error message. Fix the code and save and then `rock_on` will fix it up and keep trucking.

That's your dev cycle: save file, reload.

Now point your browser at http://127.0.0.1:8000/ and note that the title of your site has changed.  Aw, yeah!

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

## Seven5 doesn't do windows

The Seven5 server does no server side presentation handling.  Read the previous sentence again, because it means what it says and that's freaky for most web developers.  Literally, Seven5 will not render a template or change HTML in any way as it goes over the wire.

You don't need it. Really.

Seven5 serves up awesome APIs and automatically generates the Javascript to manipulate them.  All that's left is for you to write the UI in the native languages of the web: HTML, Javascript, and CSS.

Take a look at mogrel2/static/index.html and you'll notice that when the document is ready it loads the Site model and uses it to populate the title element.  Remember the Site struct mentioned above?  Well, what you're seeing is the Seven5 provided [Backbone.js](http://documentcloud.github.com/backbone/) model which fronts the Go defined data structure with the same name.

But, does that mean that the client has to fetch the Site data every time it loads a page?  Mais, non!  Seven5 also provides a way to tag a Go data structure to be cached on the client using web storage, so in fact we're doing less network traffic by sending the site name over the wire once instead of again and again in server rendered templates.

## Carve out a URL space

Ok, enough talk.  Let's build something.

When an HTTP request comes in to Seven5 it looks for a static file to load first.  Failing that, it looks at the first element in the path (like "plarst" in "/plarst/42") and looks for an HTML file with the same root, like "plarst.html".  If it finds that file then it serves it unchanged and otherwise it's a 404.

So, let's carve out all URLs under /plarst/ (like /plarst/42 or /plarst/really/long/url/) by creating blargh/mongrel2/static/plarst.html.  For the purposes of this demo, just copy index.html to plarst.html and change "Welcome" to "Bite me".

Now load http://127.0.0.1:8000/plarst/ (or http://127.0.0.1:8000/plarst/anything/really/) and note that you are indeed loading plarst.html.

## Move through space and time

If you look in plargh.html you'll notice that it creates a Backbone.js router.  This is the object that looks at the URL and decides what to render.  Let's make plargh.html show different content for different URLs.

Add this line to routes:

     "detail/:plargh":	"detail"

And add this function to the router:

    detail: function(plargh){
		$('body').empty();
		$('body').html('You are at the plargh named ' + plargh);
	}


To recap, Seven5 will serve up everything under and including /plargh/ by simply serving plargh.html.  Then plargh.html will use the Backbone.js router to render the appropriate content to the page.

So, in plargh.html you'd define different Backbone views for URLs like /plargh/43 (a detail page for a post) and /plargh/archive/.

## Enjoy the free HTML, CSS, Ajax and events

Anything which can be automated should be.  Yes, you can reach a flow state implementing yet another RESTful API and fronting it with JS objects and staying in sync via websockets.  You can also use chainsaws to carve bears statues out of stumps.  Pick a better hobby and don't do it on company time!

Seven5 provides you with the ability to define your Go data structures like normal and then signal how they are to be served by the RESTful API, wrapped by Backbone.js models, cached in web storage, and updated using web socket events.

### Creating webish data structures

> TBD: persisting in cloud ram, searching

### Service for 2<sup>32</sup>

> TBD: tagging for backbonification

### Don't make me tell you again

> TBD: tagging for cacheing

### Keep and eye out

> TBD: keep updated via websocket events

## Squirt it to the cloud

> TBD
