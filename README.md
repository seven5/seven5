<link href="http://kevinburke.bitbucket.org/markdowncss/markdown.css" rel="stylesheet"></link>
# Seven5

## An opinionated, stiff web framework written in Go and Dart

* Seven5 is RESTful with remorse or pity.
* Seven5 is fiercely static, yet all about dynamism.

## go and Dart programming with Seven5

### About this document

This document is a bit more detailed than is strictly necessary.  Many more words have been added to 
_explain_ what various commands and settings do because it is assumed that this is the first project
that the reader has worked on with these two programming languages.  If you find places where the
_why_ is not addressed, please report it as a bug.

This document assumes the reader understands web programming and has some experience with other languages
like Python, Ruby, Javascript, Java, or similar.  This document does not try to explain concepts like REST,
HTTP, and similar because it is assumed that the reader already knows these things.  Further, this document
assumes that you would prefer to use command-line tools and your own text editor instead of the graphical
tools like Eclipse.  There is an Eclipse plugin for go development and a fairly complete Dart 
programming environment implemented as an Eclipse "product".  These are beyond the scope of this document.
Finally, we assume the reader is familiar with git as a source code control tool, at least as a basic
level.

This document is not intended as a reference, it is intended to be read _in order_ because many of the parts
are inter-related.

### Get the codez

You need to [download and install go](http://golang.org/doc/install#install) and its support tools.  We are going
to be using the "traditional" go tools, not the gcc-based tools referred to in that document as `gccgo`.  You
don't need to do configuration yet, we'll do that in a minute, just verify that the program `go` is in
your `PATH`:

``` console
$ which go
/usr/local/go/bin/go
```

You also need to [download Dart](http://www.dartlang.org/docs/editor/);  the page referenced
is for downloading the dart editor but that's the easiest way to get all the tools in one download.  
We are going to be using the [Dartium](http://www.dartlang.org/dartium/) browser during development--
which is really just Chrome plus the Dart VM.  No Dart tools will be in your `PATH` yet.

### These codez

If you somehow managed to get this `README` without the sources, you need to `git clone` the seven5
project from github.  This should create a directory `seven5` with all the project source code, including
this file in `/path/to/seven5/README.md`.

### Your environment

I use a script at the command line to set my environment for each project that I work on.  For this 
project it is called `enable-seven5` and you run it in bash with `source`, rather than executing it.  Here is
the script:

``` bash
export GOPATH=/path/to/modena/go
export DART_SDK=/path/to/dart/dart-sdk
export PATH=$PATH:$GOPATH/bin:$DART_SDK/bin
```

After you create this script and `source` it in the shell, some dart commands should be in your `PATH`

``` console
$ which dart2js
/path/to/dart/dart-sdk/bin/dart2js
```
 
### GOPATH
 
`GOPATH` is the critical environment variable for using dart as it controls many aspect of the behavior of
the `go` tool.  The `go` program can do many things including compile projects, understand dependencies and
when they are "out of date", download packages from the internet, and run tests.  In summary, `GOPATH` tells
go where your personal source code repository is.  Inside that repository must be three directories,
`bin`, `pkg`, and `src`.  You'll see this structure if you look inside the `go` directory of `/path/to/modena`.

`bin` is for the built executables of your code, or code you have downloaded from the internet.  These are
placed in `bin` so it's easy to get them all with one addition to your `PATH` variable, as we did above
in our script.  `pkg` is for your, or other peoples, libraries; these are kept as `.a` files inside this 
directory although the structure underneath pkg varies based on operating system and processor that the 
library was compiled on.  

The `src` directory is the most interesting. Inside `/path/to/modena/go/src/modena` is the code for the _library_
called `modena`.  This can be thought of as all the server-side code for Modena, except for the `main()`
function to actually run it.  Since there are multiple ways to use the Modena code, there is a directory
called `cmds` that has one subdirectory for each executable needed.  One of these, for example,
is `/path/to/modena/go/src/cmds/dbload` that will load some sample data into the system database.  It should be
clear that this functionality uses some Modena code, but is quite separate from the normal operation of 
the server.

More details about `GOPATH` can be obtained with `go help gopath` from the command line.

#### Testing your GOPATH

You can start by downloading a markdown processor called [blackfriday](https://github.com/russross/blackfriday)
like this:

``` console
$ go get github.com/russross/blackfriday-tool
```

>>> That command will finish surprisingly fast.  It takes about 5.8 seconds on a home network with
    my laptop.  Of that, only 0.6 seconds was time actually doing computing, the rest is network delay.
    go is written in go.

Now, you can look inside  `/path/to/modena/go/src/github.com/russross` to see the code that was 
downloaded, compiled, and installed. There are two github projects here, `blackfriday` and 
`blackfriday-tool`.  The first is a library that allows any go program to add a fast, safe markdown processor,
and the second is a command line program that takes  various options and calls the library.  Be sure to 
note that the `go get` above just indicated the name of the tool but `go get` is smart enough to download,
compile and install dependencies as well. 

You should now be able to use the markdown processor from the command line, because `$GOPATH/bin` is in
your `PATH`:

``` console
$ which blackfriday-tool
/path/to//modena/go/bin/blackfriday-tool
$ blackfriday-tool /path/to/modena/README.md > /path/to/modena/README.html
```

The latter of those two commands processes this markdown source code in `README.md` into HTML.

#### gocode

go ships with a standard library that can parse go source code and to some extent understand go's
type system. Thus, there are many third-party tools
to do various types of things with go sources.  One of the most popular is `gocode` which you can get
the same way as above:

``` console
$ go get github.com/nsf/gocode
```
>>>> This one takes about 4.7 seconds to completely install, with 1.4 seconds of compute time.

gocode is a program that analyzes go source code and generates plausible completions based on go's
scoping rules.  This program runs as a server and code that wants a completion calls over the network
to the gocode server which analyzes the source and returns a set of completions.  This is sufficiently
fast that it can re-analyze the source code each call, even with the networking delay!

We'll return to gocode in a bit, when we talk about setting up an editor for go programming.  Most
editors (vim, textmate, emacs, eclipse) do not bother to provide go "autocomplete" as a feature
themselves because of the presence of gocode.

### Other parts of Modena

Besides the `go` subdirectory in `/path/to/modena/go` explained above, there are also the subdirectories 
`dart`, `db`, and `web`.  

* `dart` contains only the dart source code for the front-end of the application.
* `db` contains the database (sqlite3) that Modena depends on for serving content and some json files
  that provide some "starting content" for a new Modena installation.
* `web` contains html and css files needed to make the web application work in a browser. 

### Initializing a database

Let's run the tool `dbload` to initialize a database with some content.  Assuming you have already 
run the script that sets the environment variables.

``` console
$ go get cmds/dbload
$ dbload
[ some output here telling you the path to the database and what content is being loaded]
``` 
>>>> Despite dependencies on a number of github libraries that had to be downloaded, this go get
     runs in about 6.7 seconds on my laptop.

It is important to understand what just happened: using `GOPATH` the program `go` found the sources to
the program `dbload`, in `/path/to/modena/go/src/cmds/dbload/dbload.go`.  This has a number of dependencies
that you can see at the top of that source file, including the `modena` library.  That library is
present in the `GOPATH` also so it was compiled from `/path/to/modena/go/src/modena/*.go` into
`/path/to/modena/go/pkg/darwin_amd64/modena.a`.  

>>>> The segment after `pkg` varies based on operating system and processor type.

After all the dependencies are either found or downloaded, `go` compiled and built the program `dbload` into
`/path/to/modena/go/bin/dbload`.  In addition, you can see the new downloaded packages in 
`/path/to/modena/go/src/github.com`.

#### Normal use of go

With all sources downloaded with `go get`, you can now use the _local only_ version of the go 
command to build and install dbload based on your changes.  For example:

``` console
$ rm /path/to/modena/go/bin/dbload
$ go install cmds/dbload
``` 
This will create the binary again into `/path/to/modena/go/bin/dbload` and is the most common way to invoke the 
go compiler/linker and get the output put into your path.  

>>>> There is also `go build`, which does not install built artifacts, `go clean`, and `go test`.

### Running the server

Assuming you have run `dbload` above to initialize a database, usually `/path/to/modena/db/modena.sqlite`, 
you can run the server by building and running the server.  This is done with the `go` tool and the source code
in `/path/to/modena/go/src/cmds/runserver.go`.  

``` console
$ go install cmds/runserver
$ runserver
[log messages telling you which database is being used and the REST resources mapped into the server space]
``` 

### Running the debug browser

With the server running, use a different shell to start the Chromium browser; this browser has support for
Dart built-in so it allows "shift-reload" as the dev cycle for Dart.  It should be located in the directory
that is the parent of your `DART_SDK`.  On a mac, you can open this browser like this:

``` console
$ open $DART_SDK/.../Chromium.app
``` 

>>> This is probably quite different on linux.

Once you have the browser open, you can hit the server on this URL, `http://localhost:3003/static/modena.html`
which runs a very simple Dart program.  This prints a list of the quotes from `/path/to/modena/db/quotes.json`
into the browser as a list.  

#### Summary of data movement

* Use `dbload` to load the static text file `quotes.json` into tables in `modena.sqlite`
* Use `runserver` to serve up a resource that is `http://localhost:3003/quote/` as json
* Use `Chromium` to run the dart code at `hello.dart` which parses the returned json and builds HTML elements
  from it


### Developing with a text editor/command line

Most modern text editors have support for both Dart and Go, although keep in mind you typically need to
have `gocode` running for automatic completion in Go.

Editor       | Dart                                       | Go
-------------|--------------------------------------------|-------------------------------------------------
TextMate     | [Part Of Dart Source Bundle](http://code.google.com/p/dart/source/browse/branches/bleeding_edge/dart/tools/utils/textmate/) | [Via GitHub and Alan Quatermain](https://github.com/AlanQuatermain/go-tmbundle)
VIM          |[VIM scripts mirror on github](https://github.com/bartekd/vim-dart) | [VIM scripts mirror on github](https://github.com/jnwhiteh/vim-golang)
Sublime      | [Configuring Sublime Text 2 for Dart](http://active.tutsplus.com/tutorials/workflow/quick-tip-configuring-sublime-text-2-for-dart-coding/) | [Github repo of plugins](https://github.com/DisposaBoy/GoSublime)

>>>> I have not tested any of this VIM and Sublime stuff; I have no idea if it works properly.

Normally, you want to keep a shell window open that has the correct environment variables set 
for the modena project.  In this shell, you can quickly recompile the server with `go install cmds/runserver` and
restart it with `runserver`.  This is sufficiently fast that as yet I have not bothered to automate it.

When developing Dart code, two things to keep in mind.

1. Dart's `print` function goes to the debug console in Chromium, ala `console.log` in Javascript.
2. Chromium + the modena server correctly detect changed dart files so you can just "reload the page" to see updates on the client side.



>>>>>>> modena/master

