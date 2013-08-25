package seven5

import (
	"github.com/iansmith/qbs"
)

//QbsRestIndex is the QBS version of RestIndex
type QbsRestIndex interface {
	Index(PBundle, *qbs.Qbs) (interface{}, error)
}

//QbsRestFind is the QBS version of RestFind
type QbsRestFind interface {
	Find(Id, PBundle, *qbs.Qbs) (interface{}, error)
}

//QbsRestDelete is the QBS version of RestDelete
type QbsRestDelete interface {
	Delete(Id, PBundle, *qbs.Qbs) (interface{}, error)
}

//QbsRestPut is the QBS version RestPut
type QbsRestPut interface {
	Put(Id, interface{}, PBundle, *qbs.Qbs) (interface{}, error)
}

//QbsRestPost is the QBS version RestPost
type QbsRestPost interface {
	Post(interface{}, PBundle, *qbs.Qbs) (interface{}, error)
}

//QbsRestAll is the same as RestAll but with the additional qbs.Qbs parameter
//on each method.
type QbsRestAll interface {
	QbsRestIndex
	QbsRestFind
	QbsRestDelete
	QbsRestPost
	QbsRestPut
}

//qbsWrapped is just a type for wrapping around qbs-based rest methods that
//want to "appear" as simple rest methods.  Note that this is type safe and
//there is no worry about nil values if you use the QbsWrap* methods.
type qbsWrapped struct {
	store *QbsStore
	index QbsRestIndex
	find  QbsRestFind
	del   QbsRestDelete
	put   QbsRestPut
	post  QbsRestPost
}

func (self *qbsWrapped) applyPolicy(pb PBundle, fn func(*qbs.Qbs) (interface{}, error)) (result_obj interface{}, result_error error) {
	tx := self.store.Policy.StartTransaction(self.store.Q)
	defer func() {
		if x := recover(); x != nil {
			result_obj, result_error = self.store.Policy.HandlePanic(tx, x)
		}
	}()
	value, err := fn(tx)
	return self.store.Policy.HandleResult(tx, value, err)
}

//Index meets the interface RestIndex but calls the wrapped QBSRestIndex
func (self *qbsWrapped) Index(pb PBundle) (interface{}, error) {
	return self.applyPolicy(pb, func(tx *qbs.Qbs) (interface{}, error) {
		return self.index.Index(pb, tx)
	})
}

//Find meets the interface RestFind but calls the wrapped QBSRestFind
func (self *qbsWrapped) Find(id Id, pb PBundle) (interface{}, error) {
	return self.applyPolicy(pb, func(tx *qbs.Qbs) (interface{}, error) {
		return self.find.Find(id, pb, tx)
	})
}

//Delete meets the interface RestDelete but calls the wrapped QBSRestDelete
func (self *qbsWrapped) Delete(id Id, pb PBundle) (interface{}, error) {
	return self.applyPolicy(pb, func(tx *qbs.Qbs) (interface{}, error) {
		return self.del.Delete(id, pb, tx)
	})
}

//Put meets the interface RestPut but calls the wrapped QBSRestPut
func (self *qbsWrapped) Put(id Id, value interface{}, pb PBundle) (interface{}, error) {
	return self.applyPolicy(pb, func(tx *qbs.Qbs) (interface{}, error) {
		return self.put.Put(id, value, pb, tx)
	})
}

//Post meets the interface RestPost but calls the wrapped QBSRestPost
func (self *qbsWrapped) Post(value interface{}, pb PBundle) (interface{}, error) {
	return self.applyPolicy(pb, func(tx *qbs.Qbs) (interface{}, error) {
		return self.post.Post(value, pb, tx)
	})

}

//Given a QbsRestAll return a RestAl
func QbsWrapAll(a QbsRestAll, s *QbsStore) RestAll {
	return &qbsWrapped{store: s, index: a, find: a, del: a, put: a, post: a}
}

//Given a QBSRestIndex return a RestIndex
func QbsWrapIndex(indexer QbsRestIndex, s *QbsStore) RestIndex {
	return &qbsWrapped{index: indexer, store: s}
}

//Given a QbsRestFind return a RestFind
func QbsWrapFind(finder QbsRestFind, s *QbsStore) RestFind {
	return &qbsWrapped{find: finder, store: s}
}

//Given a QbsRestDelete return a RestDelete
func QbsWrapDelete(deler QbsRestDelete, s *QbsStore) RestDelete {
	return &qbsWrapped{del: deler, store: s}
}

//Given a QbsRestPut return a RestPut
func QbsWrapPut(puter QbsRestPut, s *QbsStore) RestPut {
	return &qbsWrapped{put: puter, store: s}
}

//Given a QbsRestPost return a RestPost
func QbsWrapPost(poster QbsRestPost, s *QbsStore) RestPost {
	return &qbsWrapped{post: poster, store: s}
}
