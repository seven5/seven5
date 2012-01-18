# Seven5: The Dev Loop

The following assumes that you are going to be developing in `samples` which is actually part of the *seven5* system itself.  Since *seven5* is pretty raw right now, we are assuming you don't want to do a "install and forget", since you are likely to hack at *seven5* as well as your application.

Verify that in the step above "Building Everything With gb" created two subdirectories of your `seven5dev` directory (created above in the "Seven5" step).   These two directories are `_obj` for packages and `bin` for commands.  In `_obj` for example should be `seven5.a` and others.  In bin are the seven5 tools, critically `rock_on` and `tune`.

You need to set your SEVEN5BIN to point to this directory of commands, like this:

	export SEVEN5BIN=/Users/foo/seven5dev/bin
	
This environment variable is _only_ needed because we are not installing seven5 in `$GOBIN` as we are assuming it is a bit premature for that.  This environment variable tells `rock_on` to override your `$PATH` when looking for `tune`.  When seven5 is more mature, we'll install all the tools to `$GOBIN`.

## Building A Seven5 Application

Go to the application directory for `better_chat` 
	
	cd seven5dev/samples/better_chat

and try to build with `gb`.  You should get an error like this:

	(in .) building pkg "better_chat"
	index.html.go:3: can't find import: "seven5/dsl"
	1 broken target
	(in .) could not build "better_chat"

again, this is product of not installing the *seven5* packages in the standard place, `$GOROOT/pkg`.

To remedy this problem, you need to tell `gb` where to find the "root" of your build tree, in this case your `seven5dev` directory.  Create the file `gb.cfg` in the `better_chat` directory.  It needs only this line:

	workspace=../..
	
This will allow it to find all your seven5 libraries in `seven5dev/_obj`.

### Back to normal

The normal way to do *seven5* development is to run `rock_on` in your application directory with the full name of the project's package as a parameter:
	
	cd seven5dev/samples/better_chat
	$SEVEN5BIN/rock_on samples/better_chat
	
The parameter is needed because `rock_on` cannot determine if the project's name should be "better_chat" or "samples/better_chat"  or "seven5/samples/better_chat", etc.  You should see output like this:

	Generated _seven5/better_chat.go
	Running gb in workspace /Volumes/External/seven5dev
	Cleaning seven5/dsl
	Cleaning samples/better_chat
	Cleaning mongrel2
	Cleaning seven5
	Cleaning samples/better_chat/_seven5
	(in seven5/dsl) building pkg "seven5/dsl"
	(in samples/better_chat) building pkg "samples/better_chat"
	(in mongrel2) building pkg "mongrel2"
	(in seven5) building pkg "seven5"
	(in samples/better_chat/_seven5) building cmd "better_chat"
	Built 10 targets
	Seven5 is logging to mongrel2/log/seven5.log
	Web app running
	
This means that your web application's "main" has been generated to `_seven5/better_chat.go` and the application compiled and linked with the necessary *seven5* libraries.  Plus, the web server mongrel2 has been run with your application connected to it.  Try this URL:

	http://localhost:6767/css/site.css

That is the CSS file that is generated as a result of the go-language DSL code in `site.css.go`.  Similarly:

	http://localhost:6767/index.html

is a result of `index.html.go`

`rock_on` should be left running in a shell window.  Go to another shell window and try this:

	cd seven5dev/samples/better_chat
	touch index.html.go

You should see output like this in the window running `rock_on`

	--------------------------------------
	Project samples/better_chat changed
	--------------------------------------
	Generated _seven5/better_chat.go
	Running gb in workspace /Volumes/External/seven5dev
	Cleaning seven5/dsl
	Cleaning samples/better_chat
	Cleaning mongrel2
	Cleaning seven5
	Cleaning samples/better_chat/_seven5
	(in seven5/dsl) building pkg "seven5/dsl"
	(in samples/better_chat) building pkg "samples/better_chat"
	(in mongrel2) building pkg "mongrel2"
	(in seven5) building pkg "seven5"
	(in samples/better_chat/_seven5) building cmd "better_chat"
	Built 10 targets
	Seven5 is logging to mongrel2/log/seven5.log
	Web app running

This happens automatically as you change `.go` files in your project.  This does all the work to "bring back up" all your changes to the web server as well, so you can hit `Command-R` in your web browser to reload the pages with the changes.

