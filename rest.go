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
type RestDelete interface {
	Delete(int64, PBundle) (interface{}, error)
}
type RestPut interface {
	Put(int64, interface{}, PBundle) (interface{}, error)
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

type restObj struct {
	t     reflect.Type
	name  string
	index RestIndex
	find  RestFind
	del   RestDelete
	post  RestPost
	put   RestPut
}
