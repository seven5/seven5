# Seven5: Install

<nav>
    <ul>
        <li>[Intro](index.html)</li>
        <li>[Install](install.html)</li>
        <li>[Develop](develop.html)</li>
        <li>[Pontificate](pontificate.html)</li>
    </ul>
</nav>

>In preparation for [Go1](http://blog.golang.org/2011/10/preview-of-go-version-1.html) we are setting up *Seven5* to build using the [go](http://weekly.golang.org/cmd/go/) command line tool.  This tool is not fully ready, but there is a workaround that is sufficient for us to use it.  See the discussion below about GOPATH and goinstall for more on this.

Here are the commands which we run to install the [Seven5 source](https://github.com/seven5/seven5) and prerequisites on our OS X boxen.  These install instructions expect you to be on *at least* the go weekly build of Jan 15 2012 ("weekly.2012-01-15 11253").

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
    export GOPATH=~/Documents/gothirdparty
    export PATH=$PATH:$GOBIN:$GOPATH/bin
	
## GOPATH

The recommended way ("the one true way") is to use GOPATH to keep from polluting your Go installation (at GOROOT).  You do this by keeping third party code, such as *Seven5*, in this directory.  You need to make this directory "look like" a GOPATH directory by doing this:
	
	mkdir $GOPATH/bin
	mkdir $GOPATH/src
	mkdir $GOPATH/pkg

## 0mq  : Networking library for talking to Mongrel2, if not already installed

    git clone https://github.com/zeromq/zeromq2-1.git zeromq
    cd zeromq
    ./autogen.sh
    ./configure
    make
    sudo make install

## mongrel2 : Web server of choice, if not already installled

Mongrel2 is the web server of the gods--or at least of Zed Shaw.

    curl http://mongrel2.org/static/downloads/mongrel2-1.7.5.tar.bz2 > mongrel2-1.7.5.tar.bz2
    bunzip2 mongrel2-1.7.5.tar.bz2
    tar -xvf mongrel2-1.7.5.tar
    cd mongrel2-1.7.5
    sudo make all install

## memcached, if not already installed

Download the source from [memcached](http://memcached.org/).

    tar -xzvf memcached-1.4.10.tar.gz
    cd memcached-1.4.10
    ./configure
    sudo make install

### Optional: Black Friday for markdown compilation

You only need this if you want to build the docs on your local machine.  The docs are included in the git wad-of-stuff on the branch "gh-pages".

	git clone https://github.com/russross/blackfriday.git blackfriday
	cd blackfriday
	gomake install

Copy example/markdown into /usr/local/bin/ or some other place in your $PATH.

	
## Install needed go libraries

When the "go" tool is fully finished, this step will not be needed.  For now, in a shell that has all the environment variables set as indicated above do this:

	cd $GOPATH
	go get code.google.com/p/go.crypto/bcrypt
	go get github.com/seven5/gozmq
	go get launchpad.net/gocheck
	go get github.com/bradfitz/gomemcache/memcache
	go get github.com/mattn/go-sqlite3
	go get github.com/seven5/mongrel2
	go get github.com/petar/GoLLRB/llrb

You should be able to see all the source code for these packages in $GOPATH/src and you should be able to see the packages installed in $GOPATH/darwin-amd64/ (or whatever directory gets derived from your GOOS+GOARCH). 

## Install Seven5 source

We'll install *Seven5* "the right way" using the git and the go tool.  First get the source code and put in the right place in your GOPATH:

	cd $GOPATH/src
	git clone git@github.com:seven5/seven5.git -b master seven5
	git clone git@github.com:seven5/seven5.git -b sample-dungheap seven5
	git clone git@github.com:seven5/seven5.git -b gh-pages doc
	
	go install seven5 seven5/store seven5/tune seven5/rock_on seven5/big_idea
	
## Test Seven5

When you run tests or a *Seven5* application, you have to have a "store" running.  For now, that means memcached, as you installed earlier.  In a different shell, just run "memcached" and leave it running or, if you want to get all fancy-pants "memcached -vv", so you can some feedback that *Seven5* is actually reading and writing to memcached.

Back in the shell from above where you have built *Seven5* try this:

	go test seven5 seven5/store
	
You should see the test results when this is done, something like this

	ok  	seven5	1.042s
	ok  	seven5/store	0.143s
	
	

## Optional but handy software

### TextMate bundle

    curl -L github.com/downloads/AlanQuatermain/go-tmbundle/install-go-tmbundle.sh | sh

Edit TextMate's global PATH variable to include $GOBIN.


	

	


