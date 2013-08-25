//It is convention to name your package, and thus your library, the same as the project.
package nullblog

import (
	_ "github.com/coocood/qbs"
	"github.com/seven5/seven5"
	"net/http"
	"time"
)

//ArticleWire is the name of the resource (noun) that will be exchanged over the wire.  The defined structure
//gives the content to be sent back and forth.
type ArticleWire struct {
	Id      seven5.Id
	Content seven5.String255
	Author  seven5.String255
}

//Storage into the database
type Article struct {
	Id     int64
	Author string `qbs: "size:127"`

	//this one is actually a "text" field because it has unlimited size
	Content string `qbs: "varchar"`

	// These fields are auto updated on save
	Created time.Time
	Updated time.Time
}

//ArticleResourceis the _implementation_ of the server side that exchanges ArticleWire with the
//client side.
type ArticleResource struct {
	//STATELESS! PUTTING STATE IN A REST RESOURCE CAN CAUSE THE SUN TO BURN OUT.
}

func (IGNORED *ArticleResource) Index(bundle seven5.PBundle) (interface{}, error) {
	return nil, nil
}

func (IGNORED *ArticleResource) Find(Id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
	i := int64(Id)
	if i < 0 {
		return nil, seven5.HTTPError(http.StatusBadRequest, "nice try, loser")
	}
	//err := hd.Where("color", "=", "green").OrderBy("name").Limit(1).Find(&results)
	//if err != nil {
	//	panic(err)
	//}

	return nil, nil
}

func (IGNORED *ArticleResource) Allow(IGNORED2 seven5.Id, IGNORED3 string, IGNORED4 seven5.PBundle) bool {
	return true
}

func (IGNORED *ArticleResource) AllowRead(IGNORED2 seven5.PBundle) bool {
	return true
}
