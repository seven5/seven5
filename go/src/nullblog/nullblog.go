//It is convention to name your package, and thus your library, the same as the project.
package nullblog

import (
"github.com/seven5/seven5"
	"github.com/iansmith/hood"
	"net/http"
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
	Id     hood.Id `sql:"pk" validate:"presence"`
	Author string  `sql: "size(255)" validate:"presence"`

	//this one is actually a "text" field because it has unlimited size
	Content string `validate:"presence"`

	// These fields are auto updated on save
	Created hood.Created
	Updated hood.Updated
}

//ArticleResourceis the _implementation_ of the server side that exchanges ArticleWire with the 
//client side.
type ArticleResource struct {
	//STATELESS! PUTTING STATE IN A REST RESOURCE CAN CAUSE THE SUN TO BURN OUT.
}


func (IGNORED *ArticleResource) Index(bundle seven5.PBundle) (interface{}, error) {
	return someArticle, nil
}

func (IGNORED *ArticleResource) Find(Id seven5.Id, bundle seven5.PBundle) (interface{}, error) {
	i := int64(Id)
	if i < 0  {
		return nil, seven5.HTTPError(http.StatusBadRequest, "nice try, loser")
	}
	err = hd.Where("color", "=", "green").OrderBy("name").Limit(1).Find(&results)
  if err != nil {
      panic(err)
  }
  
	return someArticle[i], nil
}

func (IGNORED *ArticleResource) Allow(IGNORED2 seven5.Id, IGNORED3 string, IGNORED4 seven5.PBundle) bool {
	return true
}

func (IGNORED *ArticleResource) AllowRead(IGNORED2 seven5.PBundle) bool {
	return true
}
