--- 
layout: book
chapter: Setup, Ninjas
---

You need to have the following tools installed on your system. With each tool, we provide a link to download the tool and read some documentation if you wish plus a comment regarding the ninja-level  required for understanding this book.


* [Git 1.8](http://git-scm.com/downloads) --- novice ninja
* [Go 1.1](http://golang.org/doc/install) --- intermediate ninja
* [Dart 0.6.14.0_r25719 (build 25719)](http://www.dartlang.org/tools/) --- powerful, but deity-level, ninja while Dart, and in particular Polymer.dart, is still stabilizing
* [Postgres 9](http://www.postgresql.org/) --- novice ninja.  Any subversion of version 9 will likely work; versions based on Postgres 8 might work.  Although out of scope for the book, the database portion of the code should be easy to port to MySQL or SQLite3 since these are directly supported by [Hood](https://github.com/eaigner/hood).

For the command line tools, these will need to be installed into your `PATH` environment variable.  These are "git", "go", "pub" and "dart".  For the inevitable database peeping, having "psql" in your `PATH` is probably a good idea too.

You should also be able to launch the Dartium browser from your Dart installation.  How to launch this browser depends on your operating system.  On a mac, the command "open $DART_SDK/chromium/Chromium.app" opens the Dartium browser.

