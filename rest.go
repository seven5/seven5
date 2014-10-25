package seven5

import (
	"reflect"
)

type RestIndex interface {
	Index(PBundle) (interface{}, error)
}
type RestFind interface {
	Find(int64, PBundle) (interface{}, error)
}
type RestFindUdid interface {
	Find(string, PBundle) (interface{}, error)
}
type RestDelete interface {
	Delete(int64, PBundle) (interface{}, error)
}
type RestDeleteUdid interface {
	Delete(string, PBundle) (interface{}, error)
}
type RestPut interface {
	Put(int64, interface{}, PBundle) (interface{}, error)
}
type RestPutUdid interface {
	Put(string, interface{}, PBundle) (interface{}, error)
}

type RestPost interface {
	Post(interface{}, PBundle) (interface{}, error)
}

type RestAll interface {
	RestIndex
	RestFind
	RestDelete
	RestPost
	RestPut
}

type RestAllUdid interface {
	RestIndex
	RestFindUdid
	RestDeleteUdid
	RestPost
	RestPutUdid
}

type restShared struct {
	t     reflect.Type
	name  string
	index RestIndex
	post  RestPost
}

type restObj struct {
	restShared
	find RestFind
	del  RestDelete
	put  RestPut
}

type restObjUdid struct {
	restShared
	find RestFindUdid
	del  RestDeleteUdid
	put  RestPutUdid
}
