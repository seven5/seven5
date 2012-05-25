package routegen

import (
	"github.com/bradrydzewski/routes"
)

type MuxTriple struct {
	Method string
	Matcher string
	FunctionName string
}

func main() {
	MuxTriple* t = &MuxTriple{"GET","/app/:id","SomeFunc(id)"};
	
	
	
}