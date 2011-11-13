Why Mongrel2
============

Production Reasons
------------------

* [Mongrel2](http://www.mongrel2.org) is derived from mongrel.  Both of them 
have extremely well tested and secure http handling code.  Both are known to
perform well under high load and to pass [valgrind](http://www.valgrind.org), so
they do not leak memory.  It's solid.

* Mongrel2 is friendly for deployment/operations and can be easily configured to
work in a cluster.  Mongrel2 can also handle having clusters, not necessarily in
the same configuration, that handle the requests for one or more applications
deployed on the cluster.  It's scalable.

Testing Reasons
---------------

* Test the "front door" not some other path.  In other words, the best tests 
use the code path that is as close---or better yet identical too--the code path 
that is used by the end-user.  Unit tests in **Seven5** code through the exact
same dispatching (sometimes called "routing") as a request in a production
deployment, even in a clustered deployment.

* Mongrel is easy to configure and control programmatically.  **Seven5** 
exploits this ability to allow the server to be configured based on its
own conventions of how to develop a web application.  During development it 
should never be necessary to touch a configuration file.  **Seven5** also uses
this ability to programmatically start or restart mongrel2 as needed to run
the developer's web application.


