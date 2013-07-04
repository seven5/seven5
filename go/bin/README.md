### Example Project For Seven5, "myproj"

#### This directory, myproj/go/bin

This directory is where your go binaries are installed, such as `runmyproj`

#### PATH

You should add this directory to your path if you want to make it easy to run the binaries that are
part of "myproj".  Here is the command to do that, assuming you have correctly set `GOPATH` as
explained in the parent directory:

```
export PATH=$PATH:$GOPATH/bin
```

#### Scripts

This directory is also the place to put scripts that do not depend on the current working directory
to execute.  Typically these are scripts that are knowledgeable enough about `GOPATH` and the Seven5
project layout to do their work independent of where they are invoked from.

#### Web mapping

The root of a web application, such as one run with "runmyproj", is mapped 
from "/" to be "myproj/dart/myproj/web".
