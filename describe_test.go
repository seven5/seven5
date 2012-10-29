package seven5

import (
	"reflect"
	"testing"
)

/*-------------------------------------------------------------------------*/
type Rec1 struct {
	Id Id
	A  Boolean
}
type Rec2 struct {
	Id Id
	S  *Rec1
	D  String255
	A  []Rec3
}

type Rec3 struct {
	Id Id
	X  Floating
	Y  Floating
}

/*-------------------------------------------------------------------------*/
/*                          VERIFICATION CODE                              */
/*-------------------------------------------------------------------------*/
func verifyNoNesting(T *testing.T, f ...*FieldDescription) {
	for _, someField := range f {
		if someField.Array != nil || someField.Struct != nil {
			T.Errorf("Unexpected nesting in field description %+v", f)
		}
	}
}
func verifyHaveCollectionAndRes(T *testing.T, d *ResourceDescription, uri string) {
	if d == nil {
		T.Errorf("Unable to generate doc for %s", uri)
	}
	if !d.Index || !d.Find {
		T.Errorf("Incorrect find or index description for %s (Index %v, Find %v)", uri,
			d.Index == true, d.Find == true)
	}
}
func verifyDocSlices(T *testing.T, d *ResourceDescription, uri string, expectedColl []string,
	expectedResource []string) {
	if d == nil {
		T.Errorf("Unable to generate doc for %s", uri)
	}
	if len(d.CollectionDoc) != len(expectedColl) {
		T.Errorf("Wrong size for collection documentation (%d vs %d)!", len(d.CollectionDoc), len(expectedColl))
	}
	for i, s := range d.CollectionDoc {
		if s != expectedColl[i] {
			T.Errorf("Unexpected collection doc on item %d: %s vs %s", i, s, expectedColl[i])
		}
	}
	for i, s := range d.ResourceDoc {
		if s != expectedResource[i] {
			T.Errorf("Unexpected resource doc on item %d: %s vs %s", i, s, expectedResource[i])
		}
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/

func TestRecursiveTraversal(T *testing.T) {
	d := WalkJsonType(reflect.TypeOf(Rec2{}))
	if len(d.Struct) != 4 {
		T.Errorf("expected 3 fields from Rec2 but found %d", len(d.Struct))
	}
	if len(d.Struct[1].Struct) != 2 {
		T.Errorf("expected 4 fields from Rec1 but found %d", len(d.Struct[1].Struct))
	}
	if d.Struct[1].StructName != "Rec1" {
		T.Errorf("Expected to find Rec1 nested in Rec2 but got %s", d.Struct[1].StructName)

	}
	if d.Struct[3].Array == nil {
		T.Errorf("expected array type, but didn't find it: %+v", d.Struct[3].Array)
	}
	verifyNoNesting(T, d.Struct[0], d.Struct[2])
	verifyNoNesting(T, d.Struct[1].Struct[0], d.Struct[1].Struct[1])

	fd := d.Struct[3].Array
	verifyNoNesting(T, fd.Struct[0], fd.Struct[1], fd.Struct[2])
}

func TestResolve(T *testing.T) {
	h := NewSimpleHandler()
	h.AddResourceByName("person", Ox{}, &oxFinder{})
	res, id, _ := h.resolve("/person/123")
	if res != "/person/" || id != "123" {
		T.Errorf("Unable to resolve /person/123 correctly (res=%s and id=%s)!", res, id)
	}
	res, id, _ = h.resolve("/person/")
	if res != "/person/" || id != "" {
		T.Errorf("Unable to resolve /person/ correctly (res=%s and id=%s)!", res, id)
	}
	res, id, _ = h.resolve("/person/456")
	if res != "/person/" || id != "456" {
		T.Errorf("Unable to resolve /person/456 correctly (res=%s and id=%s)!", res, id)
	}
}

func TestDescribe(T *testing.T) {
	h := NewSimpleHandler()
	h.AddExplicitResourceMethods("person", Ox{}, &oxIndexer{}, &oxFinder{}, nil, nil, nil)

	p := "/person/129"

	person, _, _ := h.resolve(p)
	verifyHaveCollectionAndRes(T, h.Describe(person), p)

	verifyDocSlices(T, h.Describe(person), p, []string{"FOO", "bar", "Baz"},
		[]string{"How can you lose an ox?", "fleazil", "frack for love"})

}
