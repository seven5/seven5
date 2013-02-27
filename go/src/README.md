### Example Project For Seven5, "myproj"

#### This directory

This directory will have child directories that are the names of _your_ packages, such as `myproj`.  In
addition it will have child directories like `github.com` which contain source code that you have
downloaded from the internet.  

You can use `go get github.com/seven5/seven5` to download the Seven5 source code as an example.  Typically,
`go build` will automatically do this for you as needed based on your source code's imports.

If downloading the Seven5 source code fails, please check your `GOPATH` and see the parent directory of
this one for instructions about `GOPATH`.

