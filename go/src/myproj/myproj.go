//It is convention to name your package, and thus your library, the same as the project.
package myproj

import (
	"github.com/seven5/seven5"
	"net/http"
)

//GreetingWire is a type that will be exchanged over the wire. This will be marshalled to/from Json as needed.
//It must contain Id field and all the type
type GreetingWire struct {
	Id           seven5.Id
	Greeting     seven5.String255
	LanguageName seven5.String255
}

//GreetingResource is the _implementation_ of the server side that exchanges GreetingWire with the 
//client side.
type GreetingResource struct {
	//STATELESS!
}

//see http://www.roesler-ac.de/wolfram/hello.htm
var someGreetings = []*GreetingWire{
	&GreetingWire{
		0, "Hello, World", "Amerikan",
	},
	&GreetingWire{
		1, "Saluton, mondo", "Esperanto",
	},
	&GreetingWire{
		2, "Salut le Monde", "French",
	},
	&GreetingWire{
		3, "Dia dhaoibh, a dhomhain", "Irish",
	},
}

func (IGNORED *GreetingResource) Index(bundle seven5.PBundle) (interface{}, error) {
	return someGreetings, nil
}

func (IGNORED *GreetingResource) Find(Id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
	i:=int64(Id)
	if i<0 || int(i)>len(someGreetings) {
		return nil,seven5.HTTPError(http.StatusBadRequest, "nice try, loser")	}
}
