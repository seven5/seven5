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

## The testing tool, gocheck

    bzr branch lp:gocheck
    cd gocheck
    make install


## 0mq

    git clone https://github.com/zeromq/zeromq2-1.git zeromq
    cd zeromq
    ./autogen.sh
    ./configure
    make
    sudo make install

## Gozmq

    goinstall github.com/alecthomas/gozmq

## mongrel2

    curl http://mongrel2.org/static/downloads/mongrel2-1.7.5.tar.bz2 > mongrel2-1.7.5.tar.bz2
    bunzip2 mongrel2-1.7.5.tar.bz2
    tar -xvf mongrel2-1.7.5.tar
    cd mongrel2-1.7.5
    sudo make all install

## libevent

Download the libevent source from [libevent](http://www.monkey.org/~provos/libevent/).

    tar -xzvf libevent-2.0.16-stable.tar.gz
    cd libevent-2.0.16-stable
    ./configure
    sudo make install

## memcached

Download the source from [memcached](http://memcached.org/).

    tar -xzvf memcached-1.4.10.tar.gz
    cd memcached-1.4.10
    ./configure
    sudo make install

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

### Black Friday for markdown compilation
    git clone https://github.com/russross/blackfriday.git blackfriday
    cd blackfriday
    gomake install

Copy example/markdown into /usr/local/bin/ or some other place in your $PATH.

