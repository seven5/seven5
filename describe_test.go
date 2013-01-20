package seven5

import (
	"reflect"
	"strings"
	"testing"
)

/*-------------------------------------------------------------------------*/
type Rec1 struct {
	Id	Id
	A	Boolean
}
type Rec2 struct {
	Id	Id
	S	*Rec1
	D	String255
	A	[]Rec3
}

type Rec3 struct {
	Id	Id
	X	Floating
	Y	Floating
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
func verifyHaveMethods(T *testing.T, d *APIDoc, uri string) {
	if d == nil {
		T.Errorf("Unable to generate doc for %s", uri)
	}
	if !d.Index || !d.Find || !d.Post || !d.Put || !d.Delete {
		T.Errorf("Incorrect method description for %s (Index %v, Find %v, Post %v, Put %v, Delete %v)", uri,
			d.Index == true, d.Find == true, d.Post == true, d.Put == true, d.Delete == true)
	}
}
func verifyDocStructs(T *testing.T, uri string /* basedocset or bodydocset*/, expected interface{},
	/* basedocset or bodydocset*/ actual interface{}) {

	expectedBody, ok := expected.(*BodyDocSet)

	var expectedBase *BaseDocSet
	var actualBase *BaseDocSet

	if ok {
		actualBody := actual.(*BodyDocSet)

		//test bodies
		if strings.Index(actualBody.Body, expectedBody.Body) == -1 {
			T.Errorf("can't find '%s' in actual body '%s' of %s", expectedBody.Body, actualBody.Body, uri)
		}
		if strings.Index(actualBody.Body, expectedBody.Body) == -1 {
			T.Errorf("can't find '%s' in actual body '%s' of %s", expectedBody.Body, actualBody.Body, uri)
		}
		if strings.Index(actualBody.Body, expectedBody.Body) == -1 {
			T.Errorf("can't find '%s' in actual body '%s' of %s", expectedBody.Body, actualBody.Body, uri)
		}
		if strings.Index(actualBody.Body, expectedBody.Body) == -1 {
			T.Errorf("can't find '%s' in actual body '%s' of %s", expectedBody.Body, actualBody.Body, uri)
		}
		return
	} else {
		expectedBase = expected.(*BaseDocSet)
		actualBase = actual.(*BaseDocSet)
	}

	if strings.Index(actualBase.QueryParameters, expectedBase.QueryParameters) == -1 {
		T.Errorf("can't find '%s' in actual query params '%s' of %s", expectedBase.QueryParameters,
			actualBase.QueryParameters, uri)
	}
	if strings.Index(actualBase.Headers, expectedBase.Headers) == -1 {
		T.Errorf("can't find '%s' in actual headers '%s' of %s", expectedBase.Headers, actualBase.Headers, uri)
	}
	if strings.Index(actualBase.Result, expectedBase.Result) == -1 {
		T.Errorf("can't find '%s' in actual result '%s' of %s", expectedBase.Result, actualBase.Result, uri)
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

//to make the typing easier
type oxResource struct {
	oxFinder
	oxIndexer
	oxPuter
	oxPoster
	oxDeleter
}

func TestResolve(T *testing.T) {
	T.Logf("describe_test.TestResolve needs updating!")
	/*
	h := NewSimpleHandler(nil)
	h.AddResourceByName("person", Ox{}, &oxResource{})
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
	*/
}

func TestDescribeGeneration(T *testing.T) {
	T.Logf("describe_test.TestDescribeGeneration needs updating!")
	/*h := NewSimpleHandler(nil)
	h.AddResourceByName("person", Ox{}, &oxResource{})

	p := "/person/129"

	person, _, _ := h.resolve(p)
	verifyHaveMethods(T, h.Describe(person), p)

	rd := h.Describe(person)

	if rd.FindDoc == nil {
		T.Fatalf("No find documentation found: %+v", rd)
	}

	verifyDocStructs(T, "/person/129 FIND", &BaseDocSet{
		`lose an ox`,
		`awk`,
		`sed`,
	},
		rd.FindDoc)

	verifyDocStructs(T, "/person/129 INDEX", &BaseDocSet{
		`FOO`,
		`Bar`,
		`Baz`,
	},
		rd.IndexDoc)

	verifyDocStructs(T, "/person/129 POST", &BodyDocSet{
		`Grik`,
		`Grak`,
		`Frobnitz`,

		`fleazil`,
	},
		rd.PostDoc)

	verifyDocStructs(T, "/person/129 PUT", &BodyDocSet{
		`leia`,
		`luke`,
		`assumed to be form data`,
		`isn't used`,
	},
		rd.PutDoc)

	verifyDocStructs(T, "/person/129 DELETE", &BaseDocSet{
		`mickey`,
		`minnie`,
		`pluto`,
	},
		rd.DeleteDoc)
		*/
}
