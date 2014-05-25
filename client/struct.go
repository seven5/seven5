package client

import (
	"strings"
	//"github.com/gopherjs/jquery"
	"fmt"
)

//Models represent a collection of stuff to be displayed in a view.
type ModelName interface {
	Id() string
}

type Model interface {
	ModelName
	Equaler
}

var (
	count = 0
)

type ModelImpl struct {
	n int
	t string
}

func (self *ModelImpl) Id() string {
	return fmt.Sprintf("%s-%d", self.t, self.n)
}

func NewModelName(i interface{}) ModelName {
	t := fmt.Sprintf("%T", i)
	if t[0] == '*' {
		t = t[1:]
	}
	idx := strings.Index(t, ".")
	if idx != -1 {
		t = t[idx+1:]
	}
	result := &ModelImpl{count, t}
	count++
	return result
}
