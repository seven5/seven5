package seven5

import (
	"testing"
	"strings"
)

const exampleapi1=`
{
	"Index": true,
  "Find": true,
	"CollectionURL": "/oxen/",
	"CollectionDoc": 
	[
		"FOO",
		"bar",
		"Baz"
	],
	"ResourceURL": "/ox/123",
	"ResourceDoc": 
	[
		"How can you lose an ox?",
		"fleazil",
		"frack for love"
	],
	"Fields": 
	[
		{
		   "Name": "Id",
		   "GoType": "int64",
		   "DartType": "int",
		   "SQLType": "integer",
		   "Array": null,
		   "Struct": null
		},
		{
		   "Name": "IsLarge",
		   "GoType": "bool",
		   "DartType": "bool",
		   "SQLType": "int",
		   "Array": null,
		   "Struct": null
		}
	]
}`

const exampleapi1_dart = `
class Ox {
	int Id;
	bool IsLarge;
}
`
/*-------------------------------------------------------------------------*/
/*                          VERIFICATION CODE                              */
/*-------------------------------------------------------------------------*/

func verifyHasString(T *testing.T, s string, code string) {
	if strings.Index(code,s)==-1 {
		T.Errorf("expected to find %s in the generated code:\n%s\n",s,code)
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/

func TestDartFields(T *testing.T) {
	h:=NewSimpleHandler()
	h.AddFindAndIndex("ox",&ExampleFinder_correct{},"oxen",&ExampleIndexer_correct{}, Ox{})
	
	p:="/ox/129"
	//q:="/oxen/"
	
	person, _, _ := h.resolve(p)
	//people, _, _ := h.resolve(q)
	doc:=h.GenerateDoc(person)
	
	decl:=doc.generateDart()
	verifyHasString(T,"class Ox {",decl)
	verifyHasString(T,"int Id;",decl)
	verifyHasString(T,"bool IsLarge;",decl)
	verifyHasString(T,"Ox();",decl)
	verifyHasString(T,"Ox.copyFromJson(Map json)",decl)
	verifyHasString(T,"static List<Ox> Index(",decl)
	verifyHasString(T,"void Find(",decl)
	verifyHasString(T,"static String indexURL = \"oxen\"",decl)
	verifyHasString(T,"static String findURL = \"ox\"",decl)
}
