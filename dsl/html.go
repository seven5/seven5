package dsl

import (
	"fmt"
	"strings"
)

//interface so we know who is an HTML Elem and who is not
type HtmlElem interface {
	htmlCloser() string
}

//for now we just point to the text of the .js
type JsSource string

//Document is a head and a body... plus the name at the HTTP level
type Document struct {
	Name string
	H    Head
	B    Body
}

// HTML
func (self Document) String() string {
	result := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
`
	result += fmt.Sprintf("%s\n", self.H)

	result += fmt.Sprintf("%s\n", self.B)

	return result + self.htmlCloser()
}

func (self Document) htmlCloser() string {
	return "</html>\n"
}

//HEAD
type Head struct {
	Title   string
	Sheet   *StyleSheet
	Sources []JsSource
}

func (self Head) String() string {
	t := `<title>%s</title>\n`
	s := `<link rel="stylesheet" href="css/%s" />
	`
	j := `<script src="%s" type="text/javascript"></script>
	`
	head := `<head>
	%s
	
	<link rel="stylesheet" href="/static/css/reset.css" />
    <link rel="stylesheet" href="/static/css/text.css" />
    <link rel="stylesheet" href="/static/css/960.css" />
    
	%s
	
	%s
`
	title := fmt.Sprintf(t, self.Title)
	Css := fmt.Sprintf(s, self.Sheet.Name)
	script := ""
	for _, jsName := range self.Sources {
		script += fmt.Sprintf(j, jsName)
	}
	return fmt.Sprintf(head, title, Css, script) + self.htmlCloser()
}

func (self Head) htmlCloser() string {
	return "</head>\n"
}

//BODY
type Body struct {
	R []Row
}

func (self Body) String() string {
	result := "<body><div class=\"container_12\">\n"
	for i,r :=range(self.R) {
		if !r.CheckChildrenWidth(12) {
			panic(fmt.Sprintf("failed check width in %dth row of body",i))
		}
		info:=r.Info(true)
		for _,div:=range(info) {
			result+=fmt.Sprintf("%s\n",div)
		}
	}
	return result + self.htmlCloser()

}

func (self Body) htmlCloser() string {
	return "</div><!--container 12 --></body>\n"
}
//we build these as a list of entries recursively than dump the printout at the end
type ClassIdInfo struct {
	Class []string
	Id    string
	Nested []*ClassIdInfo
}

func (self *ClassIdInfo) String() string {
	id:=""
	if self.Id!="" {
		id=fmt.Sprintf(` id="%s"`,self.Id)
	}	
	cls:=""
	space:=""
	for _,cname:=range(self.Class) {
		cls+=string(space+cname)
		space=" "
	}
	classAndId:=strings.TrimLeft(fmt.Sprintf(`class="%s"`,cls)+id," ")
	inner:=""
	for _,n:=range(self.Nested) {
		inner+=fmt.Sprintf("\n%s\n",n)
	}
	return fmt.Sprintf("<div %s>%s</div>",classAndId,inner)
}

//crucial interface for checking the grid entries for corectness and then correctly
//walking the structure to emit the grid elements
type Kids interface {
	TotalWidth() int
	Children() []Kids
	CheckChildrenWidth(int) bool
	Info(isToplevel bool) []*ClassIdInfo
}

//This is just here to know when to emit the "clear" div
type Row struct {
	B []Kids
}

func (self Row) Info(isTopLevel bool) []*ClassIdInfo {
	result := []*ClassIdInfo{}
	ch := self.Children()
	for _, k := range ch {
		info := k.Info(false)
		result = append(result, info...)
	}
	result=append(result,&ClassIdInfo{Class:[]string{"clear"}})
	return result
}

func (self Row) Children() []Kids {
	return self.B
}

//strict: all rows must be exactly expected
func (self Row) CheckChildrenWidth(expected int) bool {
	if self.TotalWidth() != expected {
		return false
	}
	for _, k := range self.Children() {
		if !k.CheckChildrenWidth(k.TotalWidth()) {
			return false
		}
	}
	return true
}

func (self Row) TotalWidth() int {
	sum := 0
	for _, b := range self.B {
		sum += b.TotalWidth()
	}
	return sum
}

//This is just here to know when to emit alpha and omega
type Column struct {
	B []Kids
}

func (self Column) Children() []Kids {
	return self.B
}
//strict: all cols must be exactly expected
func (self Column) CheckChildrenWidth(expected int) bool {
	if self.TotalWidth() != expected {
		return false
	}
	for _, k := range self.Children() {
		if !k.CheckChildrenWidth(k.TotalWidth()) {
			return false
		}
	}
	return true
}
func (self Column) TotalWidth() int {
	max := 0
	for _, b := range self.B {
		if w := b.TotalWidth(); max < w {
			max = w
		}
	}
	return max
}

func (self Column) Info(isTopLevel bool) []*ClassIdInfo {
	inner := []*ClassIdInfo{}
	
	ch := self.Children()
	for i, k := range ch {
		//fmt.Fprintf(os.Stderr, "column checking kid (%d): %v\n",i,k)
		info := k.Info(false)

		if len(ch) > 1 && !isTopLevel {
			if i == 0 {
				first := info[0]
				first.Class = append(first.Class, "alpha")
			}
			if i == len(ch)-1 {
				last := info[len(info)-1]
				last.Class = append(last.Class, "omega")
			}
		}
		inner = append(inner, info...)
	}
	
	result:= []*ClassIdInfo{}
	width:=self.TotalWidth()
	cls:=fmt.Sprintf("grid_%d",width)
	result=append(result,&ClassIdInfo{Class:[]string{cls},Nested: inner})
	return result
}

//This is where the content goes
type Box struct {
	Width, Prefix, Suffix int
	//style
	Class Class
	Id    Id
}

func (self Box) TotalWidth() int {
	return self.Width + self.Prefix + self.Suffix
}

func (self Box) Children() []Kids {
	return nil
}

func (self Box) CheckChildrenWidth(expected int) bool {
	return self.TotalWidth() == expected
}

func (self Box) Info(isTopLevel bool) []*ClassIdInfo {
	c := []string{}
	if self.Width == 0 {
		panic("all boxes have to have a width!")
	}
	c = append(c, fmt.Sprintf("grid_%d", self.Width))
	if self.Prefix != 0 {
		c=append(c, fmt.Sprintf("prefix_%d", self.Prefix))
	}
	if self.Suffix != 0 {
		c=append(c, fmt.Sprintf("suffix_%d", self.Suffix))
	}
	if self.Class != "" {
		c=append(c, fmt.Sprintf("%s", string(self.Class)))
	}
	i := string(self.Id)

	myInfo := &ClassIdInfo{c, i, nil}
	return []*ClassIdInfo{myInfo}
}

