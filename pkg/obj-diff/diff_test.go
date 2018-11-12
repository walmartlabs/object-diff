package obj_diff

import (
	"fmt"
	"reflect"
	"testing"
)

type diffTestObject struct {
	name   string
	base   interface{}
	update interface{}
	expect changeSet
}

type structInt struct {
	A int32
}

func TestDiff(t *testing.T) {
	int1 := int32(123)
	int2 := int32(456)
	i := []diffTestObject{
		buildSimpleTest("Basic -- Int", int(-123), int(123), []PathElement{}, int(123)),
		buildSimpleTest("Basic -- Uint", uint(123), uint(456), []PathElement{}, uint(456)),
		buildSimpleTest("Basic -- Float", uint(123), uint(456), []PathElement{}, uint(456)),
		buildSimpleTest("Basic -- Complex", complex128(-123.45 + 3.14i), complex128(456.78 + 2.71i), []PathElement{}, complex128(456.78 + 2.71i)),
		buildSimpleTest("Basic -- Bool", false, true, []PathElement{}, true),
		buildSimpleTest("Basic -- String", "Hello", "World", []PathElement{}, "World"),

		buildSimpleTest("Object -- Struct", structInt{A: 123}, structInt{A: 456}, []PathElement{NewNameElem(0, "A")}, int32(456)),
		buildSimpleTest("Object -- Map", map[string]int32{"A": 123}, map[string]int32{"A": 456}, []PathElement{NewKeyElem("A")}, int32(456)),
		buildSimpleTest("Object -- Array", [1]int32{123}, [1]int32{456}, []PathElement{NewIndexElem(0)}, int32(456)),
		buildSimpleTest("Object -- Slice", []int32{123}, []int32{456}, []PathElement{NewIndexElem(0)}, int32(456)),
		buildSimpleTest("Object -- Ptr", &int1, &int2, []PathElement{NewPtrElem()}, int32(456)),
	}
	tests := i
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := Diff(test.base, test.update)
			if err != nil {
				t.Fatalf("error in test: %v", err)
			}

			expect := &test.expect

			if !expect.Equals(*actual) {
				t.Logf("Not Equal:")
				t.Logf("Expect: %+v", expect)
				t.Logf("Actual: %+v", actual)
				t.Fail()
			}
		})
	}
}

func buildSimpleTest(name string, base interface{}, update interface{}, path []PathElement, newValue interface{}) diffTestObject {
	if reflect.TypeOf(base) != reflect.TypeOf(update) {
		panic(fmt.Errorf("base type (%T) != update type (%T)", base, update))
	}

	objectType := reflect.TypeOf(base)
	return diffTestObject{name: name, base: base, update: update,
		expect: changeSet{
			BaseType: objectType,
			Changes: []Change{
				{Path: path, NewValue: reflect.ValueOf(newValue), Deleted: false},
			},
		},
	}
}
