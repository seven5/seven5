--- 
layout: book
chapter: The Null Blog
---

### Let's not be clever and build a blog engine

It is _de rigeur_ to build a blogging engine if you are explaining your new, fancy web toolkit.  The author has not deviated from this well-trodden path because a blogging engine is complicated enough to present many of the problems of modern web applications, but not so tailored to a particular domain that specialized knowledge is required to understand correct operation.  This book is neither leading nor getting the hell out of the way.  The author has named this blog engine _nullblog_ because the idea is so unoriginal.

### Goal: Retreive, Build, and Run
At the end of this chapter, you should be able to get the sample code for any chapter of this book, build the source code contained in it, and run the back end of a server based on _Seven5_.  You should also be familiar with the most critical environment variables needed to configure a _Seven5_ application.

### Practice: Getting the server source

We will detail the project layout for a _Seven5_ project in the next chapter, but for now simply fetch the initial sample code, contained in branch "book\_nullblog" like this:

```
$ cd /tmp
$ git clone -b book_nullblog git@github.com:seven5/seven5.git book
```

>>>> Throughout this book will assume you are working through these in examples in the directory `/tmp` and will omit further references to it in command-lines such as the first one above.  If you are not working in `/tmp` you will need adjust some commands that use absolute paths to match your local system.

### Theory: Configuration with environment variables

_Seven5_ makes heavy use of environment variables for configuration.  This is for three reasons:

1. Using the Seven5 conventions, environment variables allow for multiple projects to developed simultaneously and to allow for easily simulation of different deployment environments and situations
2. Environment variables content are not checked into version control, avoiding a common source of errors
3. This book's deployment platform, heroku, understands environment variables as a way of configuring applications allowing the book to show the complete development process

### Theory: GOPATH is the one and only one

In the next section, we'll set the `GOPATH` environment variable.  When working on a _Seven5_ project, this should be the only configuration that references the filesystem (has a path contained in it).  The `PATH` variable references the `GOPATH` variable's "bin" directory, but should do that indirectly by using "$GOPATH/bin".

`GOPATH` should have exactly one directory in it, the go source code directory of the project being developed.  All other files inside a _Seven5_ project can have their absolute path calculated from the `GOPATH` variable, since the project layout is standardized. Experienced Go developers may complain that a `GOPATH` variable may contain many directories, and thus that this restriction is too harsh.  Experienced Go developers *may* be able to develop with multiple directories in their `GOPATH` if the _Seven5_ project is first.   This type of configuration is beyond the scope of this book but not beyond experienced Go developers.

The rationale behind the decision to derive all paths from `GOPATH` is that the "go" command requires `GOPATH` to be set in a particular way so the principle of [Don't Repeat Yourself](http://en.wikipedia.org/wiki/Don't_repeat_yourself) dictates that we do not have another environment variable repeating this value.  If there were a way to avoid repeating this value somewhat in the `PATH`, it would be used, but this is not possible in most operating systems today.

### Practice: Building and running the source code 

To test that you have gotten the source code correctly, you should set your first environment variable, [GOPATH](https://code.google.com/p/go-wiki/wiki/GOPATH) and the derived `PATH` value.  Then set your second variable, `PORT` which controls the server's port number for both testing and deployment:

```
$ export GOPATH=/tmp/book/go
$ export PATH=$PATH:$GOPATH/bin
$ export PORT=4004
```

and then test by running this command in the `/tmp/book` directory:

```
$ go get github.com/seven5/seven5
```

This command installs the _Seven5_ library code from github.  This command should complete without outputting anything, although it can take a while if you are on a slow network connection.

You can test the initial source code for the _nullblog_ application like this:

```
$ go install nullblog/runnullblog
```

>>>> When you are doing application development based on _Seven5_, a command like the one above is the one you use to "build the server".  Go is smart enough to build any dependencies that are implied in your code so you don't need a build tool, like "make".

This command is important because it compiles and installs the command `runnullblog` in your `PATH`. You can test that this built the command correctly by just invoking the command (this depends on the `PATH` modification above to connect `PATH` to `GOPATH/bin`).  

```
$ runnullblog
```

### Theory: Servers only respond to API calls

Because the `runnullblog` binary is the server portion of our application, it doesn't really _do_ anything until you connect it to a client.  You won't see any output if you run the command above, but it should not return control to the command line.  Control-c can stop it.

### Practice: Servers only respond to API calls

We won't have any type of web client for the _nullblog_ until a few chapters from now, but if you leave the command above running and use another shell, you can prove it is running with a unix command:

```
$ curl -o- localhost:4004/rest/article
[
 {
  "Id": 0,
  "Content": "This is a really short article that demonstrates the concept.",
  "Author": "Ian Smith"
 },
 {
  "Id": 1,
  "Content": "Another very short article, must be less than 255 characters!",
  "Author": "Ian Smith"
 }
]
```

You may find it interesting to try asking for a particular article with `curl -o- localhost:4004/rest/article/1` or `curl -o- localhost:4004/rest/article/0`.  

You can also point your web browser at the same URLs to see the responses, as shown in this screen capture:

<img src="https://www.evernote.com/shard/s238/sh/b80cac50-91d4-48cb-b98e-806a9678d972/e2ba585664c1a3483bd3fe73176212c3/deep/0/Screenshot%207/4/13%204:33%20PM.png"/>

>>>> You should notice that the response format from this simple _Seven5_ service is JSON not HTML.  

*Congratulations* you have a working server now and can proceed with building a better blog engine!


