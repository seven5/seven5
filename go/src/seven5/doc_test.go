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
	S  Rec1
	D  String255
	A  []Rec3
}

type Rec3 struct {
	Id Id
	X Floating
	Y Floating
}

/*-------------------------------------------------------------------------*/
/*                          VERIFICATION CODE                              */
/*-------------------------------------------------------------------------*/
func verifyNoNesting(T *testing.T, f... FieldDescription) {
	for _, someField := range f {
		if someField.Array!=nil || someField.Struct!=nil {
			T.Errorf("Unexpected nesting in field description %+v",f)
		}
	}
}
func verifyHaveCollectionAndRes(T *testing.T, d *ResourceDescription, uri string) {
	if d==nil {
		T.Errorf("Unable to generate doc for %s",uri)
	}
	if !d.Index || !d.Find {
		T.Errorf("Incorrect find or index description for %s (Index %v, Find %v)",uri,
		d.Index==true, d.Find==true)
	}
}
func verifyDocSlices(T *testing.T, d *ResourceDescription, uri string, expectedColl []string,
	expectedResource []string) {
	if d==nil {
		T.Errorf("Unable to generate doc for %s",uri)
	}
	if len(d.CollectionDoc)!=len(expectedColl) {
		T.Errorf("Wrong size for collection documentation (%d vs %d)!", len(d.CollectionDoc),len(expectedColl))
	}
	for i, s:=range d.CollectionDoc {
		if s!=expectedColl[i] {
			T.Errorf("Unexpected collection doc on item %d: %s vs %s",i,s,expectedColl[i])
		}
	}
	for i, s:=range d.ResourceDoc {
		if s!=expectedResource[i] {
			T.Errorf("Unexpected resource doc on item %d: %s vs %s",i,s,expectedResource[i])
		}
	}
}

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/

func TestRecursiveTraversal(T *testing.T) {
	d := walkJsonType(reflect.TypeOf(Rec2{}))
	if len(*d) != 4 {
		T.Errorf("expected 3 fields from Rec2 but found %d", len(*d))
	}
	if len(*(*d)[1].Struct)!=2{
		T.Errorf("expected 4 fields from Rec1 but found %d", len(*(*d)[1].Struct))
	}
	if (*d)[3].Array==nil{
		T.Errorf("expected array type, but didn't find it: %+v", (*d)[3].Array)
	}
	verifyNoNesting(T, (*d)[0], (*d)[2])
	verifyNoNesting(T, (*(*d)[1].Struct)[0], (*(*d)[1].Struct)[1])
	verifyNoNesting(T, (*(*d)[3].Array.Struct)[0],(*(*d)[3].Array.Struct)[1],(*(*d)[3].Array.Struct)[2])
}

func TestResolve(T *testing.T) {
	h:=NewSimpleHandler()
	h.AddFindAndIndex("person",&ExampleFinder_correct{},"people",&ExampleIndexer_correct{}, Ox{})
	res, id, _ := h.resolve("/person/123")
	if res!="/person/" || id!="123" {
		T.Errorf("Unable to resolve /person/123 correctly (res=%s and id=%s)!",res,id)
	}
	res, id, _ = h.resolve("/person/")
	if res!="/person/" || id!="" {
		T.Errorf("Unable to resolve /person/ correctly (res=%s and id=%s)!",res,id)
	}
	res, id, _ = h.resolve("/people/456")
	if res!="/people/" || id!="456" {
		T.Errorf("Unable to resolve /people/456 correctly (res=%s and id=%s)!",res,id)
	}
}

func TestDoc(T *testing.T) {
	h:=NewSimpleHandler()
	h.AddFindAndIndex("person",&ExampleFinder_correct{},"people",&ExampleIndexer_correct{}, Ox{})
	
	p:="/person/129"
	q:="/people/"
	
	person, _, _ := h.resolve(p)
	people, _, _ := h.resolve(q)
	verifyHaveCollectionAndRes(T,h.GenerateDoc(person),p)
	verifyHaveCollectionAndRes(T,h.GenerateDoc(people),q)
	
	verifyDocSlices(T,h.GenerateDoc(person),p,[]string{"FOO","bar","Baz"},
		[]string{"How can you lose an ox?","fleazil","frack for love"})
	verifyDocSlices(T,h.GenerateDoc(people),q,[]string{"FOO","bar","Baz"},
		[]string{"How can you lose an ox?","fleazil","frack for love"})
	
}
