package seven5

import (
	"reflect"
	"testing"
)

/*-------------------------------------------------------------------------*/
type Rec1 struct {
	Id int64
	A  bool
}
type Rec2 struct {
	Id int64
	S  *Rec1
	D  String255
	A  []Rec3
}

type Rec3 struct {
	Id int64
	X  float64
	Y  float64
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

/*-------------------------------------------------------------------------*/
/*                                 TEST CODE                               */
/*-------------------------------------------------------------------------*/

func TestRecursiveTraversal(T *testing.T) {
	d := WalkWireType("Rec2", reflect.TypeOf(Rec2{}))
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
