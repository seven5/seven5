### Example Project For Seven5 book

#### Introduction

This branch contains the sample code for a project for seven5 in the correct project layout.  This layout
is principally important for three reasons:

* Go (server-side) and Dart (client-side) have expectations about the way projects are laid out.  
Seven5 must conform to both.

* GOPATH is used to "find" not only the Go source code in the project but also all the other 
parts of the project, by assuming this particular layout.

* Deployment to Heroku in most any format requires a buildpack that understands that format.  
The _Seven5_ buildpack understands this one.

Everywhere in this project, including this directory, you should substitute the real name of your project for "nullblog". It seems like there is a `cp -R` plus `sed` + `mv` trick that should be here.

#### This Directory, "nullblog"

This top level directory is also called the Seven5 project root.  It should have two child directories for the
server and client side of the project, `go` and `dart`.

#### (Un)githubme

There are two scripts in this directory.  These are primarily for folks who are doing development of seven5
itself, or other projects that require local development but are eventually pushed to github.

`githubme.bash` is a script that looks for comments like this in go source code

```
import "seven5"//githubme:seven5:
```

When you run `githubme.bash` this is converted into:

```
import "github.com/seven5/seven5"//ungithubme:seven5:
```

Note that the "seven5" between the colons is the username of the github account and the "seven5" in quotes
is the name of the package.  This script lets you switch between local and github versions of the code
if you are editing it locally.  The file puts the original source code in a file called "sourcefile.gorepl"
but this can usually be discarded immediately.

`ungithubme.bash` is a script that undoes the actions of githubme.bash.

Both of these scripts expect `gsed` to be gnu sed.  On your system it might just be "sed".

#### Other scripts

If you have scrips that *do not* depend on the current working directory, these belong in the directory
"nullblog/go/bin".  If the scripts *do* depend on the current working directory, put them in this directory
as with `githubme.bash`.
