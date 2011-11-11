package sample
import "seven5"
type HWVue2 struct {
	seven5.SimpleVue
}
func (self *HWVue2) render(context map[string]interface{}) {
	self.Literal("hello wor"+string('l')+"d")
}
