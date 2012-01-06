# Seven5: Install

<nav>
    <ul>
        <li>[Intro](index.html)</li>
        <li>[Install](install.html)</li>
        <li>[Develop](develop.html)</li>
        <li>[Pontificate](pontificate.html)</li>
    </ul>
</nav>

Here are the commands which we run to install the [Seven5 source](https://github.com/seven5/seven5) and prerequisites on our OS X boxen.

You may need to edit the paths or config options to reflect your environment, but we made notes where that is likely.

## Go

If you lack a compiler or something else Go needs, consult the [Go install page](http://golang.org/doc/install.html).

    mkdir ~/Documents/Go/ # Anywhere which you own is fine
    cd ~/Documents/Go/
    hg clone https://go.googlecode.com/hg/ go
    cd go
    hg update weekly #gives you most recent weekly build of go
    cd src
    ./all.bash
    ./sudo.bash

To update to a new weekly:

	cd ~/Documents/Go/go/
	hg pull
	hg update weekly
	cd src
	./all.bash
	./sudo.bash

## Your shell

Some variables need to be in your environment before you proceed, so add them to ~/.bash_profile:

    export GOROOT=~/Documents/Go/go   #directory where you installed go
    export GOBIN=$GOROOT/bin
    export GOOS=darwin  #or your os, like 'linux' or 'freebsd'
    export GOARCH=amd64  #or your arch, like 'x86' for 32-bit or 'arm'
    export PATH=$PATH:$GOBIN

## The build tool, gb

    goinstall go-gb.googlecode.com/hg/gb

You'll want to make sure you have an up-to-date version of gb--we use some of the latest features of it below.

## The testing tool, gocheck

    goinstall -u launchpad.net/gocheck


## 0mq  : Networking library for talking to Mongrel2

    git clone https://github.com/zeromq/zeromq2-1.git zeromq
    cd zeromq
    ./autogen.sh
    ./configure
    make
    sudo make install

## Gozmq  : Go bindings for 0mq

    goinstall github.com/alecthomas/gozmq

There is a patch that we have been using on darwin/OS X.  It is probably not needed on 64 bit linux systems.

	From 6ea494980d21e10f555583519efe8344823dd609 Mon Sep 17 00:00:00 2001
	From: Ian Smith <iansmith@acm.org>
	Date: Tue, 22 Nov 2011 12:08:12 +0100
	Subject: [PATCH] fix another sizeof bug on darwin

	---
	 zmq.go |    2 +-
	 1 files changed, 1 insertions(+), 1 deletions(-)

	diff --git a/zmq.go b/zmq.go
	index a4225c2..be88f5c 100644
	--- a/zmq.go
	+++ b/zmq.go
	@@ -238,7 +238,7 @@ func (s *zmqSocket) destroy() {
	 // Set an int option on the socket.
	 // int zmq_setsockopt (void *s, int option, const void *optval, size_t optvallen); 
	 func (s *zmqSocket) SetSockOptInt(option IntSocketOption, value int) error {
	-       if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(&value))) != 0 {
	+       if C.zmq_setsockopt(s.s, C.int(option), unsafe.Pointer(&value), C.size_t(unsafe.Sizeof(C.int(value)))) != 0 {
	                return errno()
	        }
	        return nil
	-- 
	1.7.7.1
	

## mongrel2 : Web server of choice

    curl http://mongrel2.org/static/downloads/mongrel2-1.7.5.tar.bz2 > mongrel2-1.7.5.tar.bz2
    bunzip2 mongrel2-1.7.5.tar.bz2
    tar -xvf mongrel2-1.7.5.tar
    cd mongrel2-1.7.5
    sudo make all install

## libevent  --not needed yet but will be in the future--

Download the libevent source from [libevent](http://www.monkey.org/~provos/libevent/).

    tar -xzvf libevent-2.0.16-stable.tar.gz
    cd libevent-2.0.16-stable
    ./configure
    sudo make install

## memcached --not needed yet but will be soon--

Download the source from [memcached](http://memcached.org/).

    tar -xzvf memcached-1.4.10.tar.gz
    cd memcached-1.4.10
    ./configure
    sudo make install

### Black Friday for markdown compilation

	git clone https://github.com/russross/blackfriday.git blackfriday
	cd blackfriday
	gomake install

Copy example/markdown into /usr/local/bin/ or some other place in your $PATH.

### llrb - for balanced trees used for indexing in store

	git clone https://github.com/petar/GoLLRB.git	
	cd GoLLRB
	make install

## Seven5

Check out some repos and branches containing the mongrel2 connector, the seven5 stack, the docs and the samples.

    mkdir seven5dev
    cd seven5dev
    git clone -b master https://github.com/seven5/mongrel2 mongrel2
    git clone -b master https://github.com/seven5/seven5 seven5
    git clone -b gh-pages https://github.com/seven5/seven5 doc
    git clone -b samples https://github.com/seven5/seven5 samples

## Building Everything With gb

    cd seven5dev
    gb


## Optional but handy software

### TextMate bundle

    curl -L github.com/downloads/AlanQuatermain/go-tmbundle/install-go-tmbundle.sh | sh

Edit TextMate's global PATH variable to include $GOBIN.

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


	

	


