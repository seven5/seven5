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
type Document struct{
	Name string
	H Head
	B Body
}

// HTML
func (self Document) String() string {
	result:= `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.1//EN" "http://www.w3.org/TR/xhtml11/DTD/xhtml11.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" xml:lang="en">
`
	result+=fmt.Sprintf("%s\n",self.H)

	result+=fmt.Sprintf("%s\n",self.B)
	
	return result+self.htmlCloser()
}

func (self Document) htmlCloser() string {
	return "</html>\n"
}

//HEAD
type Head struct {
	Title string
	Sheet *StyleSheet
	Sources []JsSource
}

func (self Head) String() string {
	t:=`<title>%s</title>\n`
	s:=`<link rel="stylesheet" href="%s" />
	`
	j:=`<script src="%s" type="text/javascript"></script>
	`
	head:=`<head>
	%s
	
	<link rel="stylesheet" href="/static/css/reset.css" />
    <link rel="stylesheet" href="/static/css/text.css" />
    <link rel="stylesheet" href="/static/css/960.css" />
    
	%s
	
	%s
`	
	title:=fmt.Sprintf(t,self.Title)
	Css:=fmt.Sprintf(s,self.Sheet.Name)
	script:=""
	for _,jsName:=range self.Sources {
		script+=fmt.Sprintf(j,jsName)
	}
	return fmt.Sprintf(head,title,Css,script)+self.htmlCloser()
}

func (self Head) htmlCloser() string {
	return "</head>\n"
}


//Div element that is part of a 960 grid
type GridDiv struct {
	Width int
	Prefix int
	Suffix int
	DivId *Id
}

func (self GridDiv) String() string {
	p:=""
	s:=""
	if self.Prefix!=0 {
		p=fmt.Sprintf("prefix_%d",self.Prefix)
	}
	if self.Suffix!=0 {
		s=fmt.Sprintf("suffix_%d",self.Suffix)
	}
	if self.Width==0 {
		panic("Width must be set on a GridDiv!")
	}
	g:=fmt.Sprintf("grid_%d",self.Width)
	result:=""
	id:=""
	if self.DivId!=nil {
		idSelector:=fmt.Sprintf("%s",self.DivId)
		id=fmt.Sprintf(`id="%s"`,idSelector[1:])//clip off the #
	}
	cls:=strings.Trim(fmt.Sprintf("%s %s %s",g,p,s)," ")
	result+=fmt.Sprintf(`<div class="%s" %s>`,cls,id)
	return result+self.htmlCloser()
}

func (self GridDiv) htmlCloser() string {
	return "</div>"
}

//BODY
type Body struct {
	Content []HtmlElem
}

func (self Body) String() string {
	result:="<body><div class=\"container_12\">\n"
	for _,e:=range self.Content {
		result+=fmt.Sprintf("\t%s\n",e)
	}
	return result+self.htmlCloser()
	
}

func (self Body) htmlCloser() string {
	return "</div><!--container 12 --></body>\n"
}


