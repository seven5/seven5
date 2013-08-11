package seven5

import (
	_ "database/sql"
	_ "fmt"
	"github.com/iansmith/hood"
	"os"
	"path/filepath"
)

//HoodDBPolicy expresses policy choices about how to handle transactions and
//some errors.  Code should initialize a Policy object and then use SetHoodPolicy
//to install that policy choice(s).
var hoodPolicy HoodPolicy

func SetHoodPolicy(p HoodPolicy) {
	hoodPolicy = p
}

//HoodRestIndex is the Hood version of RestIndex
type HoodRestIndex interface {
	Index(PBundle, *hood.Hood) (interface{}, error)
}

//HoodRestFind is the Hood version of RestFind
type HoodRestFind interface {
	Find(Id, PBundle, *hood.Hood) (interface{}, error)
}

//HoodRestDelete is the Hood version of RestDelete
type HoodRestDelete interface {
	Delete(Id, PBundle, *hood.Hood) (interface{}, error)
}

//HoodRestPut is the Hood version RestPut
type HoodRestPut interface {
	Put(Id, interface{}, PBundle, *hood.Hood) (interface{}, error)
}

//HoodRestPost is the Hood version RestPost
type HoodRestPost interface {
	Post(interface{}, PBundle, *hood.Hood) (interface{}, error)
}

//HoodRestAll is the same as RestAll but with the additional hood.Hood parameter
//on each method.
type HoodRestAll interface {
	HoodRestIndex
	HoodRestFind
	HoodRestDelete
	HoodRestPost
	HoodRestPut
}

//hoodWrapped is just a type for wrapping around Hood-based rest methods that
//want to "appear" as simple rest methods.  Note that this is type safe and
//there is no worry about nil values if you use the HoodWrap* methods.
type hoodWrapped struct {
	index HoodRestIndex
	find  HoodRestFind
	del   HoodRestDelete
	put   HoodRestPut
	post  HoodRestPost
}

//Index meets the interface RestIndex but calls the wrapped HoodRestIndex
func (self *hoodWrapped) Index(pb PBundle) (interface{}, error) {
	tx := startTX()
	defer func() {
		if x := recover(); x != nil {
			hoodPolicy.handlePanic(tx, x)
		}
	}()
	value, err := self.index.Index(pb, tx)
	return hoodPolicy.HandleResult(tx, value, err)
}

//Find meets the interface RestFind but calls the wrapped HoodRestFind
func (self *hoodWrapped) Find(id Id, pb PBundle) (interface{}, error) {
	tx := startTX()
	defer func() {
		if x := recover(); x != nil {
			hoodPolicy.handlePanic(tx, x)
		}
	}()
	value, err := self.find.Find(id, pb, tx)
	return hoodPolicy.HandleResult(tx, value, err)
}

//Delete meets the interface RestDelete but calls the wrapped HoodRestDelete
func (self *hoodWrapped) Delete(id Id, pb PBundle) (interface{}, error) {
	tx := startTX()
	defer func() {
		if x := recover(); x != nil {
			hoodPolicy.handlePanic(tx, x)
		}
	}()
	value, err := self.del.Delete(id, pb, tx)
	return policy.HandleResult(tx, value, err)
}

//Put meets the interface RestPut but calls the wrapped HoodRestPut
func (self *hoodWrapped) Put(id Id, value interface{}, pb PBundle) (interface{}, error) {
	tx := startTX()
	defer func() {
		if x := recover(); x != nil {
			hoodPolicy.handlePanic(tx, x)
		}
	}()
	value, err := self.put.Put(id, value, pb, tx)
	return policy.HandleResult(tx, value, err)
}

//Post meets the interface RestPost but calls the wrapped HoodRestPost
func (self *hoodWrapped) Post(value interface{}, pb PBundle) (interface{}, error) {
	tx := startTX()
	defer func() {
		if x := recover(); x != nil {
			hoodPolicy.handlePanic(tx, x)
		}
	}()
	value, err := self.post.Post(value, pb, tx)
	return policy.HandleResult(tx, value, err)

}

//Given a HoodRestIndex return a RestIndex
func HoodWrapIndex(indexer HoodRestIndex) RestIndex {
	return &hoodWrapped{index: indexer}
}

//Given a HoodRestFind return a RestFind
func HoodWrapFind(finder HoodRestFind) RestFind {
	return &hoodWrapped{find: finder}
}

//Given a HoodRestDelete return a RestDelete
func HoodWrapDelete(deler HoodRestDelete) RestDelete {
	return &hoodWrapped{del: deler}
}

//Given a HoodRestPut return a RestPut
func HoodWrapPut(puter HoodRestPut) RestPut {
	return &hoodWrapped{put: puter}
}

//Given a HoodRestPost return a RestPost
func HoodWrapPost(poster HoodRestPost) RestPost {
	return &hoodWrapped{post: poster}
}

//HoodPolicy represents the set of policy choices made when using the Hood
//ORM, especially choices related to transactions. HandleResult is called
//in both failure and sucecss cases because it is necessary to commit
//transactions on successes.  HandlePanic is only called if a panic has
//occured inside called method.  It needs to panic again if it wants the
//panic to continue (which is usual).
type HoodPolicy interface {
	StartTransaction() *hood.Hood
	HandleResult(*hood.Hood, interface{}, error)
	HandlePanic(*hood.Hood, interface{})
}

//DefaultHoodPolicy is a simple implementation of HoodPolicy that is sufficient
//for most applications.
type DefaultHoodPolicy struct {
	hood hood.Hood
}

//NewDefaultHoodPolicy returns a new default implementation of the HoodPolicy that will Rollback
//transactions if there is a 400 or 500 returned by the client. It will also
//rollback if a non-http error is returned, or if the called code panics.  After rolling
//back the transaction, it allows the panic to continue.
func NewDefaultHoodPolicy(pf ProjectFinder, name string) (HoodPolicy, error) {
	hd, err := connectToHood(pf, name)
	if err != nil {
		return nil, err
	}
	result := &DefaultHoodPolicy{hood: hd}
	return result
}

//StartTransaction returns a new Hood object after creating the transaction.
func (self *DefaultHoodPolicy) StartTransaction() *hood.Hood {
	return self.hood.Begin()
}

//HandleResult determines whether or not the transaction provided should be rolled
//back or if it should be committed.  It rolls back when the result value is
//a non-http error, if it is an Error and the status code is >= 400.
func (self *DefaultHoodPolicy) HandleResult(tx *hood.Hood, value interface{}, err error) (interface{}, error) {
	if err != nil {
		switch e := err.(type) {
		case *Error:
			if e.StatusCode >= 400 {
				rerr := tx.Rollback()
				if rerr != nil {
					return nil, rerr
				}
			}
		}
	} else {
		if cerr := tx.Commit(); cerr != nil {
			return nil, cerr
		}
	}
	return value, err
}

//HandlePanic rolls back the transiction provided and then panics again.
func (self *DefaultHoodPolicy) HandlePanic(tx *hood.Hood, err interface{}) {
	if rerr := tx.Rollback(); rerr != nil {
		panic(rerr)
	}
	panic(err)
}

//connectToHood is a convenience method for figuring out where the hood
//configuration file is and connecting to hood based on it.
func connectToHood(pf ProjectFinder, name string) (*hood.Hood, error) {
	dbdir, err := pf.ProjectFind("db", name, TOP_LEVEL_FLAVOR)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(filepath.Join(dbdir, "config.json"))
	if err != nil {
		return nil, err
	}
	driver, source := hood.ReadConfFile(f, "development", "")

	hd, err := hood.Open(driver, source)
	if err != nil {
		return nil, err
	}

	return hd, nil
}
