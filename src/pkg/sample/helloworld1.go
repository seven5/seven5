package sample

import "seven5"

type HWVue struct {
	seven5.SimpleVue
}

func (self *HWVue) render(context map[string]interface{}) {
	self.Literal("hello world")
}
