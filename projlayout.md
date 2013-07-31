--- 
layout: book
chapter: Project layout
---

### Goal: Project Layout
After reading this chapter, you should have a familiarity with the standard _Seven5_ project layout and some understanding of the rationale for this layout. For this chapter, you can continue to use the code you checked out in the previous chapter, based on the branch "code-book-1". 

### Theory: Convention over configuration

Seven5 projects share the layout explained in this chapter for two primary reasons:

1. Having only way to lay out a project means that far fewer configuration files or variables are necessary to tell _Seven5_ where a particular sub-part of the project is.  Setting the `GOPATH` variable (see previous chapter) that points to the directory that has the standard go source code layout should be sufficient to allow development anywhere in a filesystem.

2. A single project layout structure means two _Seven5_ developers can more easily work together because each is sure where particular parts of the system "should" go.  This is also true for browsing an application's source code, written by somebody else, to determine "how did he do that"?

_Seven5_ has a more complex project layout than one might expect or prefer. This is caused largely by the fact that two programming languages are being used, and each one has standard way of laying out source packages.  It is particularly important that we follow the conventions established by the Go and Dart languages because these conventions allow a large application to be broken into packages, both on the front- and back-end.  Packages in Go can be built and managed (including packages brought in from the internet) using the `go` tool.  Packages in Dart can be managed with the `pub` tool (again included downloaded packages).

The ability to break a large application into "modules" that can be developed largely independently has proven to be an excellent feature of python's Django web toolkit, so it was summarily stolen by the author.

 _Seven5_ has a server side interface, `ProjectFinder` that is used internally to find various parts of the project "on disk" at run-time.  The `ProjectFinder` knows about Dart files, Go files, and static web files, called "assets."  _Seven5_ defaults to an implementation of `ProjectFinder` called `EnvironmentVars` that uses the `GOPATH` environment variable to find any part of the project.

### Practice: Summary of the layout

For a project named "foo":

```
// foo/
//    Procfile
//    .godir
//    dart/
//          foo/
//                web/
//                  assets/
//                      img/
//                      css/
//                pubspec.yaml
//                pubspec.lock
//                packages/
//                lib/
//                      src/
//                         app.dart
//                         ...more dart files...
//    db/
//    go/
//         bin/
//         pkg/
//         src/
//               foo/
//                  noun.go
//                  ...more go files...
//                     runfoo/
//                        main.go
//                     ...more go-based commands...
```

* The `.godir` and `Procfile` are only needed if you are deploying to heroku.  These must be used in conjunction with the _Seven5_ [buildpack](https://github.com/seven5/heroku-buildpack-go) and will be discussed in an upcoming chapter.

* The `db` directory is for configuration related to the database.

* The `go` and `dart` directories are the top level "split" between front- and back-ends.

* The layout inside `dart/foo` is standard for a dart project, including the `pubspec.yaml` and `pubspec.lock` files.

* As per the `pub` conventions, the main application code should reside in `lib/src` (private source directory) inside the dart package "foo".  This is resolved, at run-time, in Dart imports as `package:foo/src/app.dart`.  

* The default _Seven5_ behavior is to use the Dart project "foo/web" to serve _Seven5_ project "foo"; so the URL path `/assets/css/bar.css` maps to `/foo/dart/foo/web/assets/css/bar.css` in the diagram above.  Files inside the `/web/` hierarchy are assumed to be "simple files", not mapped to any code.

* The layout inside `go/` is a standard go source repository, intended for use with the `GOPATH` variable.

* _Seven5_ convention is to use `runfoo` as the name of the executable binary for a project `foo`.  This will be created by compiling `main.go`, given the above layout and typical `GOPATH`, with the command `go install foo/runfoo`.  This was shown in the previous chapter with `runnullblog`.

### Practice: README.md in the source

The source code associated with this chapter, based on the branch "code-book-1", has a `README.md` file in each directory.  This provides a detailed explanation of what code or other files belongs in a particular directory.  New _Seven5_ developers may find it useful to explore around in the source cod and read the contents of each `README.md`.
