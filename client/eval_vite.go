package client

import (
//"honnef.co/go/js/console"
)

type edgeImpl struct {
	notMarked bool //so zero value is nice
	d         node
	s         node
}

type NodeType int

const (
	NORMAL     = iota
	VALUE_ONLY = 1
	EAGER      = 2

	last = EAGER

	DEBUG_MESSAGES = false
)

var eagerQueue = []*AttributeImpl{}

//DrainEagerQueue tells eager attributes to update themselves
func DrainEagerQueue() {
	for _, a := range eagerQueue {
		a.Demand()
	}
	eagerQueue = nil
}

//newEdge returns a new edge between two given nodes.  The newly
//created edge is marked.  The source and destination attribute
//have this edge added to their respective lists, so the returned
//value is probably not that useful.
func newEdge(src node, dest node) *edgeImpl {
	e := &edgeImpl{
		s:         src,
		d:         dest,
		notMarked: false,
	}
	src.addOut(e)
	dest.addIn(e)
	return e
}

//drop edge removes the edge between two nodes.  It walks the
//edges in the outgoing list of the source node.
func dropEdge(src node, dest node) {
	if DEBUG_MESSAGES {
		print("dropping edge from ", src.id(), "->", dest.id())
	}
	src.removeOut(dest.id())
	dest.removeIn(src.id())
	dest.markDirty()
}

func (self *edgeImpl) marked() bool {
	return !self.notMarked
}

func (self *edgeImpl) mark() {
	self.notMarked = false
}

func (self *edgeImpl) clear() {
	self.notMarked = true
}

func (self *edgeImpl) src() node {
	return self.s
}

func (self *edgeImpl) dest() node {
	return self.d
}

//node is the machinery needed to implement an attribute.  It holds
//the value as an Equaler.
type node interface {
	Attribute
	walkOutgoing(fn func(*edgeImpl))
	walkIncoming(fn func(*edgeImpl))
	dirty() bool
	markDirty()
	addOut(*edgeImpl)
	addIn(*edgeImpl)
	removeOut(int)
	removeIn(int)
	source() bool
	id() int
	numEdges() int
}

//AttributeImpl is the implementation of a node.
type AttributeImpl struct {
	n            int
	name         string
	evalCount    int
	demandCount  int
	clean        bool //to make the zero value of "false" work for us
	edge         []*edgeImpl
	constraint   Constraint
	curr         Equaler
	sideEffectFn SideEffectFunc
	valueFn      ValueFunc
	nType        NodeType
}

var numNodes = 0

//ValueFunc creates values for VALUE_ONLY attributes.
type ValueFunc func() Equaler

//SideEffectFunc is one that is called AFTER a value has been
//computed to consume the value.
type SideEffectFunc func(Equaler)

//NewAttribute returns a new attribute implementation.  The node type
//should be one of NORMAL (constrainable), VALUE_ONLY (not
//constrainable) or EAGER (constrainable and always up to date).  Note
//that if you supply both funcs it is assumed that the v and s communicate
//and so the s is not called when the v produced the value.
func NewAttribute(nt NodeType, v ValueFunc, s SideEffectFunc) *AttributeImpl {
	if nt > last {
		panic("unexpected node type")
	}
	result := &AttributeImpl{
		n:            numNodes,
		nType:        nt,
		valueFn:      v,
		sideEffectFn: s,
	}
	numNodes++
	return result
}

//attach does the work of constraining this value to a set of inputs
//given by the constraint.  Note that this will panic if this node
//has a current value (meaning its a source and thus cannot be
//constrained)
func (self *AttributeImpl) Attach(c Constraint) {
	if self.nType == VALUE_ONLY {
		panic("cant attach a constraint simple value!")
	}
	if self.inDegree() != 0 {
		panic("constraint is already attached!")
	}

	self.markDirty()

	self.constraint = c
	edges := make([]*edgeImpl, len(c.Inputs()))
	for i, other := range c.Inputs() {
		edges[i] = newEdge(other.(node), self)
	}

	if self.nType == EAGER {
		//this is ok without the eager queue because we are not in
		//the process of mark ood
		self.Demand()
	}
	DrainEagerQueue()
}

//Detach removes any existing constraint dependency edges from this
//node. This method has no effect if the object has no constraint.
func (self *AttributeImpl) Detach() {
	dead := []*edgeImpl{}
	self.walkIncoming(func(e *edgeImpl) {
		e.src().removeOut(self.id())
		dead = append(dead, e)
	})
	for _, e := range dead {
		self.removeIn(e.src().id())
	}
	DrainEagerQueue()
}

func (self *AttributeImpl) SetDebugName(n string) {
	self.name = n
}

func (self *AttributeImpl) dropIthEdge(i int) {
	//console.Log("dropping ith edge", i, " from", self.id(), "size is ", len(self.edge))
	self.edge[i] = self.edge[len(self.edge)-1]
	self.edge[len(self.edge)-1] = nil
	self.edge = self.edge[0 : len(self.edge)-1]
}

//removeIn drops the pointer in the edge list to the node that has
//the id given and the receiving object as the destination of the edge.
func (self *AttributeImpl) removeIn(id int) {
	for i, e := range self.edge {
		if e.dest().id() == self.id() && e.src().id() == id {
			self.dropIthEdge(i)
			//MUST break here because the content of self.edge changed
			break
		}
	}
}

//removeOut drops the pointer in the edge list to the node that has
//the id given and the receiving object as the source of the edge.
func (self *AttributeImpl) removeOut(id int) {
	for i, e := range self.edge {
		if e.src().id() == self.id() && e.dest().id() == id {
			self.dropIthEdge(i)
			//MUST break here because the content of self.edge changed
			break
		}
	}
}

//id returns a unique number in the space of nodes (attributes).
func (self *AttributeImpl) id() int {
	return self.n
}

//walkOutgoing takes all the edges related to this node
//and selects those that start at this node and passes them
//to fn.
func (self *AttributeImpl) walkOutgoing(fn func(*edgeImpl)) {
	for _, e := range self.edge {
		if e.src().id() != self.id() {
			continue
		}
		fn(e)
	}
}

//walkIncoming takes all the edges related to this node
//and selects those that end at this node and passes them
//to fn.
func (self *AttributeImpl) walkIncoming(fn func(*edgeImpl)) {
	for _, e := range self.edge {
		if e.dest().id() != self.id() {
			continue
		}
		fn(e)
	}
}

//dirty returns true if this node's value might be out of date with
//respect to its function.
func (self *AttributeImpl) dirty() bool {
	return !self.clean
}

//markDirky set this node to dirty and marks all dependent nodes
//as dirty (the set of reachable nodes via outgoing edges).
func (self *AttributeImpl) markDirty() {
	//no sense marking something dirty twice
	if self.dirty() {
		return
	}
	self.clean = false

	//all outgoing are reachable for dirtyness
	self.walkOutgoing(func(e *edgeImpl) {
		e.dest().markDirty()
	})

	if DEBUG_MESSAGES {
		print("mark dirty ---", self.name, "is eager?", self.nType == EAGER)
	}
	if self.nType == EAGER {
		eagerQueue = append(eagerQueue, self)
	}
}

func (self *AttributeImpl) assign(newval Equaler, wantSideEffect bool) Equaler {
	//deal with any side effects of the new value, but skip it if this
	//value came from the value func (otherwise)
	if wantSideEffect && self.sideEffectFn != nil {
		self.sideEffectFn(newval)
	}

	if DEBUG_MESSAGES {
		print("assign ", self.name, "<--", newval)
	}

	//update the value
	self.curr = newval
	return newval
}

//addOut adds an edge with this node as the source.
func (self *AttributeImpl) addOut(e *edgeImpl) {
	if !e.marked() {
		panic("badly constructed edge, not marked (OUT)")
	}
	self.edge = append(self.edge, e)
}

//addIn adds an edge with this node as the destination.
func (self *AttributeImpl) addIn(e *edgeImpl) {
	if !e.marked() {
		panic("badly constructed edge, not marked (IN)")
	}
	self.edge = append(self.edge, e)
}

//source is true if the node has no incoming edges (it's
//a simple value).
func (self *AttributeImpl) source() bool {
	return self.inDegree() == 0
}

//inDegree returns the number of incoming edges on the receiver.
func (self *AttributeImpl) inDegree() int {
	inDegree := 0
	self.walkIncoming(func(e *edgeImpl) {
		inDegree++
	})
	return inDegree
}

//Set can set simple values, but only on a source (not something
//that needs computing).
func (self *AttributeImpl) SetEqualer(i Equaler) {
	if !self.source() {
		panic("can't set a value on something that is not a source!")
	}
	//no change?
	if self.curr != nil && self.curr.Equal(i) {
		return
	}

	//mark dirty on all reachable nodes
	self.markDirty()

	//mark immediate outgoing edges
	self.walkOutgoing(func(e *edgeImpl) {
		e.mark()
	})
	self.assign(i, true)
	DrainEagerQueue()
}

//Demand forces all the dependencies up to date and then
//calls the function to get the new value.
func (self *AttributeImpl) Demand() Equaler {
	self.demandCount++

	if DEBUG_MESSAGES {
		print("Demand called ", self.id(), " (", self.name, ") with dirty ", self.dirty(), "\n")
	}
	//first if we are not dirty, return our stored value
	if !self.dirty() {
		return self.curr
	}

	//when we exit, we are not dirty
	self.clean = true

	//second, we are dirty so we have to pull everybody
	//we depend on up to date
	params := make([]Equaler, self.inDegree())
	pcount := 0
	self.walkIncoming(func(e *edgeImpl) {
		params[pcount] = e.src().Demand()
		pcount++
	})

	//we now know if anybody we depend on changed
	anyMarks := false
	self.walkIncoming(func(e *edgeImpl) {
		if e.marked() {
			anyMarks = true
		}
	})

	if DEBUG_MESSAGES {
		print("inside demand ", self.id(), " (", self.name, ") anyMarks = ", anyMarks, "\n")
	}

	//if nobody actually returned a different value, then we
	//don't need to reevaluate
	if !anyMarks && self.valueFn == nil {
		return self.curr
	}

	//we have to calculate the new value
	self.evalCount++
	var newval Equaler

	if DEBUG_MESSAGES {
		print("inside demand ", self.id(), " (", self.name, ") self.valueFn = ",
			self.valueFn, ",", self.constraint, "\n")
	}

	if self.valueFn != nil && self.constraint == nil {
		newval = self.valueFn()
	} else {
		newval = self.constraint.Fn(params)
	}

	//did it change?
	if self.curr != nil && self.curr.Equal(newval) {
		return self.curr
	}
	//it did change, so we need to mark our outgoing edges
	self.walkOutgoing(func(e *edgeImpl) {
		e.mark()
	})

	return self.assign(newval, true)
}

func (self *AttributeImpl) numEdges() int {
	return len(self.edge)
}

//forEffect is a "value" that is never the same.  It can only be used
//with Sinks.
type forEffect struct {
}

//ForEffect is a singleton value that is not equal to itself.
//It's useful when you want to return a value that is never
//equal to the previous value.
var ForEffect = forEffect{}

func (self forEffect) Equal(e Equaler) bool {
	return false
}
