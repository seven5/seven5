package client

import (
	"fmt"
	"testing"
)

var (
	main = NewHtmlId("section", "main")
)

func checkCallsValue(t *testing.T, sh *AttributeImpl, n int, v bool) {
	if sh.evalCount != n {
		t.Errorf("Wrong number of calls, should be %d but was %d\n", n, sh.evalCount)
	}
	if sh.curr.(BoolEqualer).B != v {
		t.Errorf("Wrong value assigned in evaluation: expected %v\n", v)
	}
}

func TestSimpleConstraint(t *testing.T) {
	b := NewBooleanSimple(false)
	sh := newComputedAttr(&BoolEq{b})

	//its 1 not 0 because EAGER attr
	checkCallsValue(t, sh, 1, false)

	//extra demand has no effect
	sh.Demand()
	checkCallsValue(t, sh, 1, false)

	//now set the underlying value this ends up doing an update because
	//the attribute is eager
	b.Set(true)
	checkCallsValue(t, sh, 2, true)

	//this has no effect
	sh.Demand()
	checkCallsValue(t, sh, 2, true)

	//no effect, because no change
	b.Set(true)
	checkCallsValue(t, sh, 2, true)

}

func newComputedAttr(c Constraint) *AttributeImpl {
	result := NewAttribute(EAGER, nil, nil)

	//this attach causes a demand because of the EAGER above
	result.Attach(c)
	return result
}

func TestComputedConstraint(t *testing.T) {
	b := NewBooleanSimple(false)
	inv := NewBooleanInverter(b)
	sh := newComputedAttr(inv)

	//we demand the value as part of creating the attachment
	if sh.evalCount != 1 {
		t.Error("did not call the inverter function")
	}

	//demand the value again has no effect
	sh.Demand()

	if sh.evalCount != 1 {
		t.Error("unnecssary call to inverters func")
	}

	b.Set(true)
	sh.Demand()

	if sh.evalCount != 2 {
		t.Error("did not call the inverters function after dirty")
	}
}

type chainAttr struct {
	*AttributeImpl
}

func (self *chainAttr) Value() bool {
	return self.Demand().(BoolEqualer).B
}

func (self *chainAttr) Set(b bool) {
	panic("cant set this value! it's computed!")
}

func newChainAttr() BooleanAttribute {
	p := new(chainAttr)
	p.AttributeImpl = NewAttribute(NORMAL, nil, nil)
	return p
}

func BenchmarkAttributeEvaluation(bench *testing.B) {
	benchmarks(false, bench)
}
func BenchmarkMarkOOD(bench *testing.B) {
	benchmarks(true, bench)
}

func benchmarks(resetAtMarkOOD bool, bench *testing.B) {
	b := NewBooleanSimple(false)
	prev := BooleanAttribute(b)
	ch := make([]BooleanAttribute, bench.N)
	for i := 0; i < len(ch); i++ {
		ch[i] = newChainAttr()
		ch[i].Attach(NewBooleanInverter(prev))
		prev = ch[i]
	}

	if resetAtMarkOOD {
		bench.ResetTimer()
	}

	b.Set(true)

	if resetAtMarkOOD {
		return
	} else {
		bench.ResetTimer()
	}

	prev.Value()
}

type testAddAttr struct {
	*AttributeImpl
}

func (self *testAddAttr) Set(i int) {
	panic("cant set the value! one way constraints, yo!")
}

func (self *testAddAttr) Value() int {
	return self.Demand().(IntEqualer).I
}

func newAddAttr(a1, a2 IntegerAttribute) IntegerAttribute {
	result := &testAddAttr{NewAttribute(NORMAL, nil, nil)}
	result.Attach(AdditionConstraint(a1, a2, func(i int, j int) int {
		return i + j + 101
	}))
	return result
}

func checkCount(t *testing.T, msg string, prev, curr int) {
	if prev != curr {
		t.Errorf("unexpect count for %s (%d vs %d)", msg, prev, curr)
	}
}

func TestIntegerConstraint(t *testing.T) {

	i1 := NewIntegerSimple(13)
	i2 := NewIntegerSimple(17)

	add := newAddAttr(i1, i2)
	if add.Value() != 101+13+17 {
		t.Errorf("expected to get 131 but got %d", add.Value())
	}

	i1.Set(2)
	if add.Value() != 101+2+17 {
		t.Errorf("expected to get 120 but got %d", add.Value())
	}

	d1 := i1.(*IntegerSimple).demandCount
	d2 := i2.(*IntegerSimple).demandCount

	dadd := add.(*testAddAttr).demandCount
	eadd := add.(*testAddAttr).evalCount

	add.Value() //doesnt even ask for the values, nothing has changed
	checkCount(t, "num demands of i1", d1, i1.(*IntegerSimple).demandCount)
	checkCount(t, "num demands of i2", d2, i2.(*IntegerSimple).demandCount)

	//number of demands of add increased, number of evals increased
	checkCount(t, "number of demands of add", dadd+1, add.(*testAddAttr).demandCount)
	checkCount(t, "num evals of addition", eadd, add.(*testAddAttr).evalCount)

	//force an update by setting a source value
	i1.Set(11)
	checkCount(t, "num evals of add", eadd, add.(*testAddAttr).evalCount)

	if add.Value() != 101+11+17 {
		t.Errorf("constraint is not correct, expected 129 but got %d", add.Value())
	}
	checkCount(t, "num evals of add", eadd+1, add.(*testAddAttr).evalCount)
	checkCount(t, "num demands of i1", d1+1, i1.(*IntegerSimple).demandCount)

	add.Value() //no effect
	checkCount(t, "num demands of i1", d1+1, i1.(*IntegerSimple).demandCount)

}

type IntValue struct {
	*AttributeImpl
}

func (self *IntValue) Set(i int) {
	panic("can't set an IntValue! Use a constraint!")
}

func (self *IntValue) Value() int {
	return self.Demand().(IntEqualer).I
}

func TestDetach(t *testing.T) {
	num := make([]IntegerAttribute, 10)
	for i := 0; i < 10; i++ {
		num[i] = NewIntegerSimple(i)
	}
	s := &IntValue{NewAttribute(NORMAL, nil, nil)}
	p := &IntValue{NewAttribute(NORMAL, nil, nil)}

	sum := SumConstraint(num...)
	prod := ProductConstraint(num[1:]...)

	s.Attach(sum)
	p.Attach(prod)

	expectedSum := 0 + 1 + 2 + 3 + 4 + 5 + 6 + 7 + 8 + 9
	expectedProd := 1 * 2 * 3 * 4 * 5 * 6 * 7 * 8 * 9

	if s.Value() != expectedSum {
		t.Errorf("bad sum value, expected %d but got %+v", expectedSum, s)
	}
	if p.Value() != expectedProd {
		t.Errorf("bad product value, expected %d but got %+v", expectedProd, p)
	}

	s.Detach()
	num[2].Set(20) // try to force an update
	if s.dirty() {
		t.Errorf("should not have any effect after constraint removed!")
	}

	//still get right value?
	expectedProd = 1 * 20 * 3 * 4 * 5 * 6 * 7 * 8 * 9
	if p.Value() != expectedProd {
		t.Error("after changing a value, the product is wrong (%d)", p.Value())
	}

	//check edge lists
	for i, attr := range num {
		n := attr.(node)
		if i == 0 && n.numEdges() != 0 {
			t.Errorf("wrong number of edges for zero (%d)", n.numEdges())
		}
		if i != 0 && n.numEdges() != 1 {
			t.Errorf("wrong number of edges for %d (%d)", i, n.numEdges())
		}
	}

	one := &IntValue{NewAttribute(NORMAL, nil, nil)}
	one.Attach(AdditionConstraint(num[0], num[1], nil))

	if num[0].(node).numEdges() != 1 || num[1].(node).numEdges() != 2 {
		t.Errorf("added more deges but they were not found on first or second node!")
	}
	p.Detach()

	if one.Value() != 1 {
		fmt.Errorf("failed to compute constraint, got %d", one.Value())
	}

	for i, no := range num {
		n := no.(node)
		if i < 2 && n.numEdges() != 1 {
			fmt.Errorf("still have constraint on first two %d", n.numEdges())
		}
		if i >= 2 && n.numEdges() != 0 {
			fmt.Errorf("should have removed all edges now, but %d has %d", i, n.numEdges())
		}
	}
}
