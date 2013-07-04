//It is convention to name your package, and thus your library, the same as the project.
package nullblog

import (
	"github.com/seven5/seven5"
	"net/http"
)

//ArticleWire is the name of the resource (noun) that will be exchanged over the wire.  The defined structure
//gives the content to be sent back and forth.
type ArticleWire struct {
	Id      seven5.Id
	Content seven5.String255
	Author  seven5.String255
}

//ArticleResourceis the _implementation_ of the server side that exchanges ArticleWire with the 
//client side.
type ArticleResource struct {
	//STATELESS! PUTTING STATE IN A REST RESOURCE CAN CAUSE THE SUN TO BURN OUT.
}

//some sample values that will let you test the server side more easily
var someArticle = []*ArticleWire{
	&ArticleWire{
		0, "This is a really short article that demonstrates the concept.", "Ian Smith",
	},
	&ArticleWire{
		1, "Another very short article, must be less than 255 characters!", "Ian Smith",
	},
}

func (IGNORED *ArticleResource) Index(bundle seven5.PBundle) (interface{}, error) {
	return someArticle, nil
}

func (IGNORED *ArticleResource) Find(Id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
	i := int64(Id)
	if i < 0 || int(i) > len(someArticle) {
		return nil, seven5.HTTPError(http.StatusBadRequest, "nice try, loser")
	}
	return someArticle[i],nil
}
