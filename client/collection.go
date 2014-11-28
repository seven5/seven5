package client

import ()

//"honnef.co/go/js/console"

//Collection is a collection models.  Note that the type of the collection
//elements are Model which implies Equaler.
//TODO: Convert to Collection and collectionImpl
type Collection struct {
	coll *AttributeImpl
	join []Joiner
}

//Joiners connect a list to something that will manipulate elements of
//the list. Typically this is used for connecting models to views.
type Joiner interface {
	Add(int, Model)
	Remove(int, Model)
}

//EqList is a convenience for talking about the contents of the collection.
//You can use collection.AttributeImpl.(EqList) to read the contents of
//the list.
type EqList []Model

//Equal handles comparison of our list to another collection. We define
//equality to be is same length and has the same contents. Empty collections
//are not equal to anything.
func (self EqList) Equal(o Equaler) bool {
	if o == nil {
		return false
	}
	other := o.(EqList)
	if len(other) == 0 {
		return false
	}
	curr := []Model(self)
	if len(curr) == 0 {
		return false
	}
	if len(curr) != len(other) {
		return false
	}
	for i, e := range other {
		if !curr[i].Equal(e) {
			return false
		}
	}
	return true
}

//computedEmpty represents just the attribute value for the empty state.
type computedEmpty struct {
	*AttributeImpl
	coll *Collection
}

//empty is private because clients of the List should be using
//EmptyAttribute()
func (self *Collection) empty() Equaler {
	if self.coll.Demand() == nil || len(self.coll.Demand().(EqList)) == 0 {
		return BoolEqualer{true}
	}
	return BoolEqualer{false}
}

func (self *computedEmpty) Set(b bool) {
	panic("can't set the value of a computed attribute (empty of the list)")
}

func (self *computedEmpty) Value() bool {
	self.clean = true //tricky
	return self.coll.empty().(BoolEqualer).B
}

//EmptyAttribute should be used by callers to assess if the list is
//is empty or not. It is a BooleanAttribute so it can be used in
//constraint functions.
func (self *Collection) EmptyAttribute() BooleanAttribute {
	result := &computedEmpty{
		NewAttribute(VALUE_ONLY, self.empty, nil), self,
	}
	//console.Log("emtpy attr is ", result.id())
	newEdge(self.coll, result)
	//console.Log("introducing edge", self.id(), "to", result.id())
	return result
}

//computedEmpty represents just the attribute value for the empty state.
type computedLength struct {
	*AttributeImpl
	coll *Collection
}

func (self *computedLength) Set(i int) {
	panic("can't set the value of a computed attribute (length)")
}

func (self *computedLength) Value() int {
	return self.coll.length().(IntEqualer).I
}

//length is a private function; use LengthAttribute() to access the
//the length.
func (self *Collection) length() Equaler {
	raw := self.coll.Demand()
	if raw == nil {
		return IntEqualer{I: 0}
	}
	return IntEqualer{len(raw.(EqList))}
}

//LengthAttribute should be used by callers to assess how many items
//are in the list. The return value is an attribute that can be used
//in constraint functions.
func (self *Collection) LengthAttribute() IntegerAttribute {
	result := &computedLength{
		NewAttribute(VALUE_ONLY, self.length, nil), self,
	}
	return result
}

//PushRaw the way to push a Model into the list _without_ having
//any checking done about the value of that model.  Note that most
//callers would probably prefer Add() that checks to see if the item
//is already in the list before doing the addition.  Note this does
//imply updating of the empty and length attributes. If the collection
//has joiners, it is called after the PushRaw has completed.
func (self *Collection) PushRaw(m Model) {
	current := self.coll.Demand()
	var result EqList

	if current == nil {
		result = EqList{m}
	} else {
		result = append(current.(EqList), m)
	}
	self.coll.SetEqualer(result)

	if self.join != nil {
		for _, j := range self.join {
			j.Add(len(result), m)
		}
	}
}

//PopRaw is the way to access the last node of the list without any
//checking.  This will panic if the list is empty.  This implies an
//update to the length and empty attribute.  If there are Joiners
//they are called after the PopRaw completes.
func (self *Collection) PopRaw() Model {
	if self.coll.Demand() == nil || len(self.coll.Demand().(EqList)) == 0 {
		panic("can't pop from a empty ListNode!")
	}

	obj := self.coll.Demand().(EqList)
	self.coll.markDirty()

	result := obj[len(obj)-1]
	if len(obj) == 1 {
		self.coll.SetEqualer(nil)
	} else {
		self.coll.SetEqualer(obj[:len(obj)-1])
	}
	if self.join != nil {
		for _, j := range self.join {
			j.Remove(len(obj)-1, result)
		}
	}

	return result
}

//NewCollection returns an empty Collection.  You should a supply a joiner here
//if you want to run a transform on insert or remove from thelist.  You can
//pass nil if you don't need any joiner at the point you create the
//collection.
func NewList(joiner Joiner) *Collection {
	result := &Collection{
		coll: NewAttribute(VALUE_ONLY, nil, nil),
	}
	if joiner != nil {
		result.join = []Joiner{joiner}
	}
	return result
}

//Add checks to see that the model is not already in the collection then
//adds it if it is not. If the object is already in the collection, this has
//no effect (but does request the value of  all the items in the list).
//If the list has any joiners, they are notified about the new element, but
//after the change has taken place.
func (self *Collection) Add(m Model) {
	//console.Log("adding an attribute %O", m)
	if self.coll.Demand() != nil {

		obj := self.coll.Demand().(EqList)
		for _, e := range obj {
			if e.Equal(m) {
				return
			}
		}
	}
	self.PushRaw(m)
}

//Remove checks to see that the model is in the collection and
//then removes it.  If the element is in the collection multiple times, only
//the first one is removed.  Calling this on an empty collection is useless
//but not an error. If the list has joiners, they are notified about
//the removal of the element but after the removal
//has occured. The object that was removed is supplied to the Joiner.
func (self *Collection) Remove(m Model) {
	obj := self.coll.Demand().(EqList)

	for i, e := range obj {
		//console.Log("demand checking %O %O", e, m)
		if e.Equal(m) {
			//last?
			if i == len(obj)-1 {
				self.PopRaw()
			} else {
				copy(obj[i:], obj[i+1:])
				obj[len(obj)-1] = nil
				self.coll.SetEqualer(EqList(obj[:len(obj)-1]))
				if self.join != nil {
					for _, j := range self.join {
						j.Remove(i, m)
					}
				}
			}
			break
		}
	}
}

//All returns all the models in this collection.  This is a copy
//of the objects internal data.
func (self *Collection) All() []Model {
	if self.coll.Demand() == nil {
		return nil
	}
	curr := self.coll.Demand().(EqList)
	result := make([]Model, len(curr))
	for i, c := range curr {
		result[i] = c
	}
	return result
}

//Puller is a function that extracts a particular attribute from a
//a model.
type Puller func(Model) Attribute

//FoldingIterator is a function of the previous value and the next
//input attribute. The latter is extracted via a Puller function, so
//one can think of this second parameter as a model.  This function
//returns two values, the first of which is passed to this function
//again on all but the last iteration.  On the last iteration, the
//2nd value is the final result.
type FoldingIterator func(interface{}, Equaler) (interface{}, Equaler)

//AllFold creates a constraint that depends on the same
//attribute in _every_ model in the collection.   The attribute to be computed
//is the first parameter. The initial value of the iterative folding
//is the second argument, and fed to the Folder on the first iteration.
//The puller is used whenever the models in the collection change to
//extract a particular Attribute that the constraint depends on.
func (self *Collection) AllFold(
	targ Attribute,
	initial interface{},
	puller Puller,
	folder FoldingIterator,
	empty Equaler) Constraint {

	result := &foldedConstraint{
		fn:      folder,
		initial: initial,
		pull:    puller,
		targ:    targ,
		empty:   empty,
	}

	self.join = append(self.join, result)
	targ.Attach(result)
	return result

}

type foldedConstraint struct {
	deps    []Attribute
	fn      FoldingIterator
	pull    Puller
	initial interface{}
	targ    Attribute
	empty   Equaler
}

func (self *foldedConstraint) Inputs() []Attribute {
	return self.deps
}

func (self *foldedConstraint) Fn(in []Equaler) Equaler {
	if len(self.deps) == 0 {
		return self.empty
	}

	//normal case
	prev := self.initial
	var e Equaler
	for _, i := range in {
		prev, e = self.fn(prev, i)
	}
	return e
}

func (self *foldedConstraint) Remove(i int, m Model) {
	dropEdge(self.deps[i].(node), self.targ.(node))
	self.targ.(node).markDirty()

	if len(self.deps)-1 == i {
		self.deps = self.deps[:len(self.deps)-1]
	} else {
		self.deps = append(self.deps[:i], self.deps[i+1:]...)
	}
	DrainEagerQueue()
}

func (self *foldedConstraint) Add(i int, m Model) {
	a := self.pull(m)
	self.deps = append(self.deps, a)
	newEdge(a.(node), self.targ.(node))
	self.targ.(node).markDirty()

	DrainEagerQueue()
}
