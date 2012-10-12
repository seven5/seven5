<link href="http://kevinburke.bitbucket.org/markdowncss/markdown.css" rel="stylesheet"></link>
# Seven5

## An opinionated, stiff web framework written in Go and Dart

* Seven5 is RESTful with remorse or pity.
* Seven5 is fiercly reactionary towards the forces of dynamism.

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
Finally, we assume the reader is familiar with git as a source code control tool, at least at a basic
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
project *examples* from github.  

``` console
cd /tmp #can be anywhere
git clone -b examples git clone -b examples git@github.com:seven5/seven5.git examples
```

This should create a directory `examples` with all the project source code, including this file in `/path/to/examples/example1/README.md`.

### Your environment

I use a script at the command line to set my environment for each project that I work on.  For this 
project it is called `enable-seven5` and you run it in bash with `source`, rather than executing it.  Here is
the script:

``` bash
export GOPATH=/path/to/examples/example1/go
export DART_SDK=/path/to/dart/dart-sdk
export PATH=$PATH:$GOPATH/bin:$DART_SDK/bin
```

After you create this script and `source` it in the shell, some dart commands should be in your `PATH`

``` console
$ which dart2js
/path/to/dart/dart-sdk/bin/dart2js
```
 
### GOPATH
 
`GOPATH` is the critical environment variable for using go and seven5 as it controls many aspect of the behavior of the `go` tool and helps the library know where your projects are.  The `go` program can do 
many things including compile projects, understand dependencies and when they are "out of date", download 
packages from the internet, and run tests.  In simplest terms, `GOPATH` tells
go where your personal, go source code repository is.  Inside that repository must be three directories,
`bin`, `pkg`, and `src`.  You'll see this structure if you look inside the `go` directory of `/path/to/example/example1/go`.

`bin` is for the built executables, either executables from your code, or code you have downloaded from the internet.  These are
placed in `bin` so it's easy to get them all with one addition to your `PATH` variable, as we did above
in our script.  `pkg` is for your, or other peoples', libraries; these are kept as `.a` files inside this 
directory although the structure underneath `pkg` varies based on operating system and processor that the 
library was compiled on.  

The `src` directory is the most interesting. Inside `/path/to/examples/example1/go/src/italy` is the code for the quite simple _library_ called `italy`.  This is the server-side code for the first example,
except for the `main()` function to actually run it. The main function is in `/path/to/examples/example1/go/src/italy/runexample1` in the file `main.go`.  The names of executables are derived from their _directories_ so the command you'll use to start a server for example1 is `runexample1`.

More details about `GOPATH` can be obtained with `go help gopath` from the command line.

#### Testing your GOPATH

You can start by getting the seven5 support tool, called [seven5tool](https://github.com/seven5/seven5tool)
like this:

``` console
$ go get github.com/seven5/seven5tool
```

>>> That command will finish surprisingly fast.  It takes about 5.8 seconds on a home network with
    my laptop.  Of that, only 0.6 seconds was time actually doing computing, the rest is network delay.
    go is written in go.

Now, you can look inside  `/path/to/examples/example1/go/src/github.com/seven5` to see the code that was 
downloaded, compiled, and installed. There are two github projects here, `seven5` and 
`seven5tool`.  The first is the seven5 library and the latter is the seven5tool that exposes a few
commands and uses the library.  Be sure to note that the `go get` above just indicated the name of the _tool_ but `go get` is smart enough to download, compile and install dependencies as well. 

You should now be able to use the seven5tool from the command line, because `$GOPATH/bin` is in
your `PATH`:

``` console
$ which seven5tool
/path/to/examples/example1/go/bin/seven5tool
$ seven5tool help
seven5tool subcommands
----------------------
help       this list
...
```

You may to try running the tests of seven5 or example1 like this:

``` console
$ go test github.com/seven5/seven5
ok  	github.com/seven5/seven5	0.043s
$ go test italy
ok  	italy	0.041s

```
#### gocode

go ships with a standard library that can parse go source code and to some extent understand go's
type system. Thus, there are many third-party tools
to do various types of things with go _sources_.  One of the most popular is `gocode` which you can get
the same way as above:

``` console
$ go get github.com/nsf/gocode
```
>>>> This one takes about 4.7 seconds to completely install, with 1.4 seconds of compute time.

gocode is a program that analyzes go source code and generates plausible completions based on go's
scoping and import rules.  This program runs as a server and code that wants a completion calls 
over the network to the gocode server which analyzes the source and returns a set of completions.  
This is sufficiently fast that it can re-analyze the source code each call, even with the networking delay!

We'll return to gocode in a bit, when we talk about setting up an editor for go programming.  Most
editors (vim, textmate, emacs, eclipse) do not bother to provide go "autocomplete" as a feature
themselves because of the presence of gocode.

### Project layout

Besides the `go` subdirectory in `/path/to/examples/example1/go` explained above, there are also the subdirectories `dart`, and `static`.  

* `dart` contains your dart source code for the front-end of the application.
* `static` contains html and css files needed to make the web application work in a browser. Files in `static` should be static.

#### Normal use of go

With all sources downloaded with `go get`, you can now use the _local only_ version of the go 
command to build and install programs or libraries based on your changes.  For example:

``` console
$ rm /path/to/examples/example1/go/bin/seven5tool
$ go install github.com/seven5/seven5tool
``` 
This will (re)create the binary (again) into `/path/to/examples/example1/go/bin/seven5tool` and is the most
common way to invoke the go compiler/linker and get the output put into your path.  Note that `go install`
will not try to download things from the internet; `go install` does not generally care about your current
directory, it derives the source locations and so forth from `GOPATH`.

>>>> There is also `go build`, which does not install built artifacts, `go clean`, and `go test`.

### Running the server

Assuming you have set things up as above,
you can run the server by building and running the program `runexample1`. `go install` _does_ check 
if there are _local_ dependencies that need building, so the command below both builds the executable `runexample1` and builds the library `italy.a` since `runexample1` uses that library.

``` console
$ go install italy/runexample1
$ runexample1
``` 

### Running the debug browser

With the server running, use a different shell to start the Chromium browser as explained above; 
this browser has support for Dart built-in so it allows "shift-reload" as the dev cycle for Dart.  
It should be located in the directory that is the parent of your `DART_SDK`.  
On a mac, you can open this browser like this:

``` console
$ DART_FLAG='--enable_type_checks --enable_asserts' open /path/to/dart/Chromium.app
``` 

The command above runs Dartium in checked mode so asserts work and you have full type-checking love.

>>> This is probably quite different on linux or windows.

Once you have the browser open, you can hit the server on this URL, `http://localhost:3003/static/italy.html`
which runs a very simple Dart program.  You will see "Hello" in the browser but most of the action is in
the javascript (sic) console where the test output is printed:

``` console
city length is 3 undefined:1
number of cities returned from Index: 3 
    city returned from Index(): [0] Turin 
    city returned from Index(): [1] Milan 
    city returned from Index(): [2] Genoa 
result of 'Index' (GET): 200 OK 
city returned from Find() was Turin with Id 0 
object that was found was Instance of 'ItalianCity' 
result of 'Find' (GET) was 200 OK 
city length is 1 
city length is 2 
city length is 1 
city length is 0 
Failed to load resource: the server responded with a status of 400 (Bad Request) http://localhost:3003/italiancities/?max=two
Failed to load resource: the server responded with a status of 400 (Bad Request) http://localhost:3003/italiancity/16
``` 

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
for the your project.  In this shell, you can quickly recompile the server with `go install` of _mylibrary/myexecutable_
and restart it with _myexecutable_.  
This is sufficiently fast that as yet I have not bothered to automate it.
