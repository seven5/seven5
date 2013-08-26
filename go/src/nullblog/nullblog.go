//It is convention to name your package, and thus your library, the same as the project.
package nullblog

import (
	"database/sql"
	"fmt"
	"github.com/iansmith/qbs"
	"github.com/seven5/seven5"
	"net/http"
	"time"
)

//ArticleWire is the name of the resource (noun) that will be exchanged over the wire.  The defined structure
//gives the content to be sent back and forth.
type ArticleWire struct {
	Id      seven5.Id
	Content seven5.Textblob
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

//NullBlogMigrate is to allow us to automatically move the database into
//the right states, particularly for testing.
type Migrate struct {
}

func (self *Migrate) Migrate_0001(m seven5.Migrator) error {
	return m.CreateTableIfNotExists(&Article{})
}

func (self *Migrate) Migrate_0001_Rollback(m seven5.Migrator) error {
	return m.DropTableIfExists(&Article{})
}

//ArticleResourceis the _implementation_ of the server side that exchanges ArticleWire with the
//client side.
type ArticleResource struct {
	//STATELESS! PUTTING STATE IN A REST RESOURCE CAN CAUSE THE SUN TO BURN OUT.
}

func (self *Article) ToWire() *ArticleWire {
	return &ArticleWire{
		Id:      seven5.Id(self.Id),
		Content: seven5.Textblob(self.Content),
		Author:  seven5.String255(self.Author),
	}
}

func (IGNORED *ArticleResource) IndexQbs(bundle seven5.PBundle, q *qbs.Qbs) (interface{}, error) {
	var articles []*Article
	if err := q.FindAll(&articles); err != nil {
		return nil, err
	}
	result := []*ArticleWire{}
	for _, a := range articles {
		result = append(result, a.ToWire())
	}
	return result, nil
}

func (IGNORED *ArticleResource) FindQbs(Id seven5.Id, bundle seven5.PBundle, q *qbs.Qbs) (interface{}, error) {
	i := int64(Id)
	if i < 0 {
		return nil, seven5.HTTPError(http.StatusBadRequest, "nice try, loser")
	}
	result := &Article{Id: i}
	if err := q.WhereEqual("Id", i).Find(result); err != nil {
		if err == sql.ErrNoRows {
			return nil, seven5.HTTPError(http.StatusBadRequest, fmt.Sprintf("bad id %d", i))
		}
		return nil, err
	}
	return result.ToWire(), nil
}

func (IGNORED *ArticleResource) Allow(IGNORED2 seven5.Id, IGNORED3 string, IGNORED4 seven5.PBundle) bool {
	return true
}

func (IGNORED *ArticleResource) AllowRead(IGNORED2 seven5.PBundle) bool {
	return true
}
