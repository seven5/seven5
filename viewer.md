--- 
layout: book
chapter: Viewing A Blog Article
---

### Goal: A Basic Front End
At the end of this chapter, you should know run a simple user interface, written in Dart, when it is connected to a _Seven5_ back-end. You should also understand the basics of how to find the key parts of a client-side "application." For this chapter, you can continue to use the code you checked out in a previous chapter, based on the branch "code-book-1". 

### Practice: Setting up and downloading the Dart libraries

```
$ cd /tmp/book/dart/nullblog
$ pub install
Resolving dependencies........................
Downloading csslib 0.4.7+7 from hosted...
...
Dependencies installed!
```

Depending on when you are reading this book, the particular libraries and their versions may differ from the above.  You can see the result of this command by looking at the large amount of code downloaded into `/tmp/book/dart/nullblog/packages`.  You will also notice that a symbolic link, also called "packages", is placed in your project in the `/tmp/book/dart/nullblog/web/` directory as well as in other places.  The details of the [pub](http://pub.dartlang.org/) command are beyond the scope of this book.

### Practice: Showing the blog

If you run the server-side on your local workstation (as explained in chapter 1) with the command `runnullblog`, you can then run the client-side by launching Dartium (see the [setup](setup.html) chapter, this varies by operating system) and going to the URL `http://localhost:4004/article.html`.  You should see a display similar to this in Dartium:

![View Blog Screen Snap](https://www.evernote.com/shard/s238/sh/7c3da0cd-a1c1-44ea-b0da-3b924b46fb11/1d6ffb6bcc22d1cd5844732a6c5c7121/deep/0/View%20Blog.png])

>>>> In a later chapter we will detail how to configure your client-side application so it works other, modern browsers.  

### Practice: Static files

As explained previously, the directory `/tmp/nullblog/dart/nullblog/web` is mapped to `/` in the URL space (notably with the exception of `/rest`) and the nullblog application serves static content from the former to web clients as the latter.  Thus, the file `/tmp/nullblog/dart/nullblog/web/article.html` is the "page" being displayed by our client-side application in the previous section.  Crucial here to understand is that because Dartium has support for the Dart language built-in the following code at the bottom of the `article.html` file kicks off the client program that lives in the file `/tmp/nullblog/dart/nullblog/web/articleapp.dart`:

```
<script type="application/dart" src="articleapp.dart"></script>
```

### Theory: Just hit reload

Dartium has been chosen for this exposition precisely because it supports Dart natively and no extra steps are necessary to "just run it"--other reloading the page.  All the dart code found in any directory below `/tmp/nullblog` is "hot", meaning it is reloaded on each page view.   Technologies, notably the now-sunsetted [web-ui](http://www.dartlang.org/docs/tutorials/web-ui/), have been deliberately avoided if they prevent this reloading behavior.  This is particularly important for changes to the visual design of an application that requires constant "tweaking" to "see how it looks".

### Theory: Code generation linking server and client

A key element of the _Seven5_ approach to [Don't Repeat Yourself](http://en.wikipedia.org/wiki/Don't_repeat_yourself) is to generate code to keep the server and client "in sync".  The creation of the `ArticleWire` type previously, as well as the discussion of nouns and verbs [in the chapter on resources](http://localhost:4000/seven5/resources.html), may make more sense if one considers that `/tmp/nullblog/dart/nullblog/lib/src/nullblog.dart` is machine-generated.  The code in this file is _derived_ from the Go type `ArticleWire` (don't edit the Dart file! edit the Go file!) and provides access to the fields defined `ArticleWire` and the methods of `ArticleResource` (see [chapter 3](resources.html)).  

This file defines a Dart class called `article` which holds the same data as the Go type `articleWire` and the Dart class `articleResource` that provides access to Go's `ArticleResource`. Note the lower-case `A` starting both names.  

```
    Future<article> articleResource.find(1);
```

The dart code above returns, roughly, an object of type `article` via the find() method on the Go type `ArticleResource`, passing in the identifier value of 1.  Behind the scenes, this actually makes a network call to the server, "GET /rest/article/1".  Since that code takes some non-zero amount of time and can fail, we return a [Dart Future](http://www.dartlang.org/articles/futures-and-error-handling/).  The intent is to allow the Dart developer to work at the level of "article" and remain mostly unaware of the networking required. 

The lower case version of the name, and the omission of the "Wire" portion of the name from the Go code, was intentional and may be mistake.  It was chosen this way to specifically "look funny" since Dart's convention is to capitalize class names.  The intent was to be extremely clear about the fact tha 

### Practice: Code generation at startup (nullblog.dart)

When you run the `nullblog` application on the command line you will see this line of output

```
seven5: generating source code for article(ArticleWire) 
```

This confirms that the server "discovered" the `ArticleWire` type and has generated the needed Dart code.  It is placed in `/tmp/nullblog/dart/nullblog/lib/src/nullblog.dart` because that directory is intended for "private" source code, not code exposed to other applications/libraries.

>>> The discovery of wire types is not automated at this time, although it could be.  Currently, the wire types mentioned in the `main()` function's call to `resourceSeparate` are the only one's discovered.

### Practice: File naming conventions, or "that's a lot of files for such a simple app!"

The nullblog client-side application has only the functionality displayed above, yet has 4 Dart files and an HTML file! This application is small, but has been crafted to cover many simple situations that a developer may find him or herself in.  

All filenames are lower case, to avoid problems with non case-sensitive file systems.  This application's client-side files are:

* article.html: This is the main entry point for displaying the articles in the blog.  The convention is to use the singular name of the resource (noun), without the "Wire" suffix as the base name for this html file if the purpose of the file is to display some set of objects of this type. It is common to refer to this as "app" in the sense that this is a complete bit of functionality for displaying articles.  A real blog engine, thus, would consist of many "apps" in this sense.

* articleapp.dart: This is the "main" (entry-point) for the functionality in `article.html`.  It is a convention to suffix the page name with "app" to indicate that this is the entry point.  It is not advised to use "article.dart" although this seems compelling because of the confusion that could result with the _machine-generated_ type `article`.  This file should not have any functionality in it; it exists entirely to bootstrap the "production" or "non-test" code-path.  This file is deliberately placed in `/tmp/book/nullblog/dart/nullblog/web/articleapp.dart` to distinguish from the code discussed below that is shared between production and testing.

* articlepage.dart: This is the code that controls the large scale structure the display in `article.html` or the article "app".  It is suffixed with "page" to indicate that it controls page-level functionality, such as how many articles are displayed.  This file creates the class `ArticlePage` which holds a reference to the type `articleResource`.  This reference is the network connection between client and server, customarily called `rez`. This code is visible to tests and placed in `/tmp/book/nullblog/dart/nullblog/lib/src`.

* articlediv.dart: This is the code for displaying a particular article (one) in HTML.  A more complex app than our current `article.html` may have many such pieces.  These should be placed in `/tmp/book/nullblog/dart/nullblog/lib/src` and should generally have a one-to-one mapping from filenames to class names.  In the current implementation, there is no executable code in this file, only declarations.

* uisemantics.dart: This file is for shared functionality across several "apps" such as a blog engine!  This is not strictly needed for `article.html` but is shown for completeness.  It currently contains just code to display a warning message to the user about the network not being available.  It is convenient to centralize these shared bits of code for re-use and testing.  This file can be found in `/tmp/book/nullblog/dart/nullblog/lib/src`.

It is not recommended that any files in a _Seven5_ application depend on case-sensitivity of filenames or include spaces in filenames.  As it is the Dart custom, Dart filenames may contain underscores but it is discouraged elsewhere as this "suggests" Dart code.  The use of dashes in filenames should be restricted as to HTML files as this is also an established custom. 

### Theory: pub and the pubspec.yaml file

Currently, the [pubspec.yaml](http://pub.dartlang.org/doc/pubspec.html) file in `/tmp/book/nullblog/dart/pubspec.yaml` that is consumed by the `pub` application is a violation of the "no configuration" policy of _Seven5_.  It is not clear how to avoid this problem at the present time.

