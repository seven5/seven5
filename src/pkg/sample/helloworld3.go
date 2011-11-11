package sample
import "seven5"
type HWVue3 struct {
	seven5.SimpleVue
}
func (self *HWVue3) render(context map[string]interface{}) {

	self.Fragment("page_header.html")
	self.Markdown("long_bloated_intro.md")
	
	//add some CSS machinery to the literal result
	self.Literal("hello world",".helloclass","#helloid")
	
	self.Fragment("page_footer.html")
}

