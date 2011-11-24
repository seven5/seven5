
See also, the [Documentation Hub](http://seven5.github.com/seven5)

Design
------------

This is the seven5 layer. Seven5 provides a "rough passthrough" for implementations that need access to the raw mongrel2 layer.  These two passthroughs are the interfaces `Httpified` and `Jsonified`.  These interfaces allow you to process messages that are very close to the raw
mongrel2 protocol--just a few things were already parsed out rather than expecting every
developer to do that.  The `Httpified` and `Jsonified` interfaces must be implemented by the developer. See `echo_rawhttp.go` and `chat_jsonservice.go` for examples.


Installation
------------


Go
--

<pre>

hg clone https://go.googlecode.com/hg/ go
cd go
hg update weekly #gives you most recent weekly build of go
cd src
./all.bash
./sudo.bash

</pre>

Your shell
----------

<pre>
	
# Add this to ~/.bash_profile
export GOROOT=~/Documents/Go/go   #directory where you installed go
export GOBIN=$GOROOT/bin
export GOOS=darwin  #or your os, like 'linux' or 'freebsd'
export GOARCH=amd64  #or your arch, like 'x86' for 32-bit or 'arm'

#these need to be in your environment before you proceed

</pre>

Build Tool (if you prefer godag)
-------------------------

<pre>

hg clone https://godag.googlecode.com/hg/ godag
cd godag
6g _gdmk.go 
6l -o gdmk _gdmk.6 
./gdmk install

</pre>

Build Tool (if you prefer gb)
-------------------------

<pre>


</pre>


0mq
---

<pre>
	
git clone https://github.com/zeromq/zeromq2-1.git zeromq
cd zeromq
./autogen.sh
./configure
make
sudo make install

</pre>

Gozmq
-----

<pre>

goinstall github.com/alecthomas/gozmq

</pre>

mongrel2
--------

<pre>
	
curl http://mongrel2.org/static/downloads/mongrel2-1.7.5.tar.bz2 > mongrel2-1.7.5.tar.bz2
bunzip2 mongrel2-1.7.5.tar.bz2
tar -xvf mongrel2-1.7.5.tar
cd mongrel2-1.7.5
sudo make all install

</pre>

seven5
------


<pre>

# Make four subdirs from 2 repos (3 branches in seven5) for development setup

mkdir seven5dev
cd seven5dev
git clone -b master https://github.com/seven5/mongrel2 mongrel2
git clone -b master https://github.com/seven5/seven5 seven5
git clone -b gh-pages https://github.com/seven5/seven5 doc
git clone -b samples https://github.com/seven5/seven5 samples

</pre>

Building Everything With gb
------

<pre>

#go to directory seven5dev
gb

</pre>
