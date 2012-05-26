package util

import (
	"container/list"
	"bytes"
	"fmt"
)

//BetterList implements extra methods to make life easier.
type BetterList struct {
	*list.List
}

//BetterList is just a wrapper around *list.List
func NewBetterList() *BetterList {
	return &BetterList{list.New()}
}

//Contains is really the missing method on *list.List.
func (self *BetterList) Contains(v interface{}) bool {
	for e := self.List.Front(); e != nil; e = e.Next() {
		if e.Value==v {
			return true
		}
	}
	return false
}

//RemoveValue removes ANY instance of this value in the list. Returs
//true if any were found
func (self *BetterList) RemoveValue(v interface{}) bool {
	found := false
	for e := self.List.Front(); e != nil; e = e.Next() {
		if e.Value==v {
			found = true
			self.List.Remove(e)
		}
	}
	return found
}

//AllValues returns a string with all the values in the list.
func (self *BetterList) AllValues() string {
	var buffer bytes.Buffer
	buffer.WriteString("[");
	for e := self.List.Front(); e != nil; e = e.Next() {
		s:=fmt.Sprintf("%v",e.Value)
		buffer.WriteString(s)
		if e != self.List.Back() {
			buffer.WriteString(", ")
		}
	}
	buffer.WriteString("]");
	return buffer.String()
}
