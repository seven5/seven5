### Example Project For Seven5 book

#### This directory, "nullblog/go/src"

This directory should be first in your GOPATH or the only directory in your GOPATH when doing development
with seven5.  This directory is used to find all the other parts of the project.

This directory has three subdirectories, `src` for source code, `bin` for executables, and `pkg` for 
compiled libraries from go source.  This arrangement is expected for go projects and allows 

* `go install nullblog` to correctly build `pkg/<os>_<arch>/nullblog.a` to build the project's go code.

* `go install nullblog/runnullblog` to correctly build `bin/runnullblog` which is your server executable.

If you have trouble with these commands not working properly, see the instructions in "nullblog/go" about
setting your `GOPATH` environment variable.