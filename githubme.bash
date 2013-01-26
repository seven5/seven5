#/bin/bash
find . -name \*.go -exec gsed --in-place=repl --expression='s/"\(.*\)"\/\/githubme:\(.*\):/"github.com\/\2\/\1"\/\/ungithub/' {} \;
go test seven5

 