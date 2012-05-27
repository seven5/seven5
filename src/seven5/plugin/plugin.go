package plugin

import (

)

type PluginSet struct {
	Validator ProjectValidator
}

var Seven5App = &PluginSet{}

func (self *PluginSet) SetProjectValidator(pv ProjectValidator) {
	self.Validator=pv
}
