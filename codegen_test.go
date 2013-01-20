package seven5

import (
	"testing"
	"strings"
	"reflect"
)

type Nested struct {
	Id	Id
	I	Integer
}

type Test1 struct {
	F	Floating
	A	[]Integer
	S	*Nested
}

/*-------------------------------------------------------------------------*/
/*                          VERIFICATION CODE                              */
/*-------------------------------------------------------------------------*/

func verifyHasString(T *testing.T, s string, code string) {
	if strings.Index(code, s) == -1 {
		T.Errorf("expected to find %s in the generated code:\n%s\n", s, code)
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/
func TestDartFields(T *testing.T) {
	f := WalkJsonType(reflect.TypeOf(Test1{}))

	if f.HasId() {
		T.Errorf("Test1 should not be a resource, no Id field!")
	}
	if !f.Struct[2].HasId() {
		T.Errorf("Nested should be a resource, it has Id field!")
	}
	verifyHasString(T, "double", f.Struct[0].Dart())
	verifyHasString(T, "List<int>", f.Struct[1].Dart())
	verifyHasString(T, "Nested", f.Struct[2].Dart())

}

func TestStructCollection(T *testing.T) {
	f := WalkJsonType(reflect.TypeOf(Test1{}))
	d := collectStructs(f)
	if len(d) != 2 {
		T.Fatalf("Expected to find %d structs but found %d", 2, len(d))
	}
	if d[0].StructName != "Test1" {
		T.Errorf("Expected to find structs %s but found %v", "Test1", d[0].StructName)
	}
	if d[1].StructName != "Nested" {
		T.Errorf("Expected to find structs %s but found %v", "Nested", d[1].StructName)
	}
}

func TestDartFullResource(T *testing.T) {
	T.Logf("codegen_test.TestDartFullResource needs updating!")
	/*
	h := NewSimpleHandler(nil)
	h.AddResource(Ox{}, &oxFinder{})
	
	p := "/ox/129"

	person, _, _ := h.resolve(p)
	//people, _, _ := h.resolve(q)
	doc := h.Describe(person)

	decl := generateDartForResource(doc)
	verifyHasString(T, "class Ox {", decl)
	verifyHasString(T, "int Id;", decl)
	verifyHasString(T, "bool IsLarge;", decl)
	verifyHasString(T, "Ox();", decl)
	verifyHasString(T, "Ox.fromJson(Map json)", decl)
	verifyHasString(T, "void Find(", decl)
	verifyHasString(T, "static String resourceURL = \"/ox/\"", decl)
	*/
}
