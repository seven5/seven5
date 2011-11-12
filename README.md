Current Dev Setup
=================

* Eclipse with [goclipse](http://code.google.com/p/goclipse/).  
* [Weekly](http://golang.org/doc/devel/weekly.html) drops of 
[go](http://golang.org).
* [Mongrel 2](https://github.com/zedshaw/mongrel2) version 1.7.5
* [Gocode](https://github.com/nsf/gocode) has to be regularly updated to keep
up with the weeklies.  This is usually just a pull as the author keeps it up
to date.  The version that ships with goclipse will **not** work with the 
weeklies.
* [0mq](http://www.zeromq.org/) version 2.1.10
* [gozmq](https://github.com/alecthomas/gozmq) but needs to be patched
to work with the weeklies. Contact me for the patch.

Some Knowlege From An Email
===========================

<pre>

From: iansmith
To: trevorfsmith

I use eclipse-indigo-java + goclipse.  Goclipse is at the stage like the early 
java versions of eclipse--you have to sometimes hit "Project > Clean" to help 
it clear its mind about what compiler errors you have.  There are a few other 
niggling problems with goclipse--I'm not thrilled with the default directory 
layout for example--but the code completion and automatic compile of the whole 
project make it ok with me.  Lots of goers, though, still just use Make and 
some prefer the go build tool gb, https://github.com/skelterjohn/go-gb

I had to install ggdb 7.3 (not gdb, ggdb!) to get a working debugger for go 
because OSX uses 6.10 or some ancient version.  I use it from the command line 
and it's pretty minimal.  I prefer to debug with fmt.Printf() anyway.

To run tests I use "make test" from the command line.  goclipse, for me at 
least, doesn't do this right and it's sufficiently fast on the command line 
that I did not investigate.

Do you know about http://godashboard.appspot.com/package  ??  It is updated 
automatically when you use goinstall--but I frankly prefer to build from source 
because I use the weeklies.  The "gofix" tool is invaluable when taking some 
random guy's code and trying to get it to work with the weekly.  
Rob Pike is da man.

http://golang.org/doc/install.html#releases  has the info on how to use hg 
to pull the updates and get the source to the current weekly.  Immediately 
after that, I pull and rebuild gocode :-)

These days the "version" of eclipse doesn't matter so much because of the 
p2 subsystem (slow as a DEAD dog--now labelled "Eclipse Marketplace" or some 
crap like that) which can handle provisioning operations and "figure out what 
you need".  If you care, I think I downloaded the "java developer" version.

I use eclipse for everything--I know you despise this.  I use the mylyn 
connector for github to allow me to see/change issues inside eclipse.  This 
combined with mylyn actually can help your focus--IMHO--although it is painful 
at first.  You can use the issue tracking without mylyn.

I cannot find a decent markdown helper/mode for eclipse.  The one I have is
dodgy.

I also use the git team provider for dealing with most git issues... it is 
called "egit".  The git model of working copies doesn't work super-well with 
the eclipse model of "workspaces."  Eclipse does not like source code repos. After 
about an hour of trying, I eventually just gave up and ended up with two 
"projects" in my workspace--neither of which is in the workspace directory.  I 
use the egit stuff in eclipse to manage where my origin is, do commits, do 
pushes, etc.  I this type of stuff through the GUI, typically with the 
perspective Git Repository Browsing.  If I am doing something I'm less familiar 
with, I use the command line until I'm sure I understand.

My two projects in eclipse are seven5 and seven5doc.  These are different git 
clones of the repo, but pointed at different branches.  When I commit to one 
I push to github repo, which is the union of these.  github knows to rebuild
web pages when you change the branch gh-pages.

</pre>