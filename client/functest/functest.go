package main

import (
	s5 "github.com/seven5/seven5/client"
)

func TestData(dom s5.NarrowDom, needInit bool) {
	d := dom.Data("foo")
	if d != "" {
		panic("before any code has run, no data expected")
	}
	dom.SetData("foo", "bar")
	d = dom.Data("foo")
	if d != "bar" {
		panic("expected bar after set")
	}
	dom.RemoveData("foo")
	d = dom.Data("foo")
	if d != "" {
		panic("after remove, no data expected")
	}
}

func TestCss(dom s5.NarrowDom, needInit bool) {
	if needInit {
		//simulate a couple of browser settings
		dom.SetCss("color", "rgb(0, 0, 0)")
		dom.SetCss("float", "none")
	}
	//need to pick a property that is NOT defined
	d := dom.Css("color")
	if d != "rgb(0, 0, 0)" {
		panic("before any code has run, no css color expected")
	}
	dom.SetCss("float", "right")
	d = dom.Css("float")
	if d != "right" {
		panic("expected to get right for float")
	}
	dom.SetCss("float", "none")
}

func TestText(dom s5.NarrowDom, needInit bool) {
	if needInit {
		dom.SetText("something")
	}
	d := dom.Text()
	if d != "something" {
		panic("failed to get starting value right")
	}
	dom.SetText("different")
	d = dom.Text()
	if d != "different" {
		panic("failed to change text value")
	}
}

func TestRadio(dom s5.NarrowDom, needInit bool) {
	if needInit {
		return
	}
	v := dom.RadioButton("suess")
	print("radio ", v)
	if v != "" {
		panic("should not have any radio button selected yet")
	}
}

func main() {
	for _, mode := range []bool{false, true} {
		s5.TestMode = mode
		elem := s5.NewHtmlId("h3", "test1")
		dom := elem.Dom()
		TestData(dom, mode)
		TestCss(dom, mode)
		TestText(dom, mode)

		r := s5.NewRadioGroup("suess")
		dom = r.Dom()
		TestRadio(dom, mode)
	}
	print("all tests passed")
}
