package sample
import "seven5"
type HWVue4 struct {
	seven5.SimpleVue
}
func (self *HWVue4) render(context map[string]interface{}) {
	self.Literal("hello world...")
	
	if context["user"]!=nil {
		self.Literal("...and welcome back!")
	}
}


