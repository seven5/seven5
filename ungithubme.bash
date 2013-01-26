#/bin/bash
find . -name \*.go -exec gsed --in-place=repl --expression='s|"github.com/\([^/]*\)/\([^"]*\)"//ungithub|"\2"//githubme:\1:|' {} \;
go install seven5/auth
go test seven5

 