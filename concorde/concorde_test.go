package concorde

import (
	"testing"
)

var (
	myid    = NewHtmlId("div", "myid")
	frik    = NewHtmlId("div", "frik")
	wumpus  = NewHtmlId("div", "wumpus")
	itsdark = NewHtmlId("A", "itsdark")

	foo    = NewCssClass("foo")
	bar    = NewCssClass("bar")
	baz    = NewCssClass("baz")
	smelly = NewCssClass("smelly")
)

func TestBasicNoPanic(t *testing.T) {

	//none of these should panic
	DIV(myid)
	DIV(myid, foo)
	DIV(myid, foo, bar)
	DIV(foo, bar)
	DIV(myid, foo, bar)
	DIV(foo, bar)
	DIV()
}

func expectPanic(fn func()) (result bool) {
	result = false
	defer func() {
		if r := recover(); r != nil {
			result = true
		}
	}()
	fn()
	return
}
func TestBasicPanic(t *testing.T) {

	if !expectPanic(func() {
		DIV(foo, bar, myid)
	}) {
		t.Errorf("expected panic when classes are before id")
	}
	if !expectPanic(func() {
		DIV(nil)
	}) {
		t.Errorf("expected panic with a nil value")
	}
	if !expectPanic(func() {
		A(myid)
	}) {
		t.Errorf("expected panic with mismatched tag")
	}
}

func TestConcordeStruct(t *testing.T) {
	div := DIV(myid)
	if div.id != "myid" {
		t.Error("id not created properly: %s", div.id)
	}

	div = DIV(foo, bar)
	if len(div.classes) != 2 {
		t.Error("classes not created properly: %d", len(div.classes))
	}
	if div.classes[1] != "bar" {
		t.Error("classes not created properly: %v", div.classes)
	}
}

func TestConcordeNested(t *testing.T) {
	div := DIV(myid, foo, baz,
		DIV(frik),
		DIV(wumpus),
		DIV(smelly,
			A(itsdark),
		),
	)

	if len(div.children) != 3 {
		t.Errorf("children not built correctly: %d", len(div.children))
	}

	if len(div.children[2].children) != 1 {
		t.Errorf("nested children not built correctly: %d", len(div.children[2].children))
	}
	if len(div.children[1].children) != 0 {
		t.Errorf("unexpected children not built correctly: %d", len(div.children[1].children))
	}
}
