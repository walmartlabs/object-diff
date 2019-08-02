// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	"fmt"
	. "github.com/walmartlabs/object-diff/pkg/obj_diff/helpers"
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"testing"
)

type diffTestObject struct {
	name   string
	base   interface{}
	update interface{}
	expect ChangeSet
}

func TestDiff(t *testing.T) {
	int1 := int32(123)
	int2 := int32(456)
	quantity1 := resource.MustParse("500Mi")
	quantity2 := resource.MustParse("1.5Gi")
	i := []diffTestObject{
		buildSimpleTest("Basic -- Int", int(-123), int(123), []PathElement{}, int(-123), int(123)),
		buildSimpleTest("Basic -- Uint", uint(123), uint(456), []PathElement{}, uint(123), uint(456)),
		buildSimpleTest("Basic -- Float", float64(3.14159), float64(2.71), []PathElement{}, float64(3.14159), float64(2.71)),
		buildSimpleTest("Basic -- Complex", complex128(-123.45+3.14i), complex128(456.78+2.71i), []PathElement{}, complex128(-123.45+3.14i), complex128(456.78+2.71i)),
		buildSimpleTest("Basic -- Bool", false, true, []PathElement{}, false, true),
		buildSimpleTest("Basic -- String", "Hello", "World", []PathElement{}, "Hello", "World"),

		buildSimpleTest("Object -- Struct", structInt32{A: int1}, structInt32{A: int2}, []PathElement{NewFieldElem(0, "A")}, int1, int2),
		buildSimpleTest("Object -- Map", map[string]int32{"A": int1}, map[string]int32{"A": int2}, []PathElement{NewKeyElem("A")}, int1, int2),
		buildSimpleTest("Object -- Array", [1]int32{int1}, [1]int32{int2}, []PathElement{NewIndexElem(0)}, int1, int2),
		buildSimpleTest("Object -- Slice", []int32{int1}, []int32{int2}, []PathElement{NewIndexElem(0)}, int1, int2),
		buildSimpleTest("Object -- Ptr", &int1, &int2, []PathElement{NewPtrElem()}, int1, int2),

		buildSimpleTest("Object -- resource.Quantity 1", &quantity1, &quantity2, []PathElement{NewPtrElem()}, quantity1, quantity2),
		buildSimpleTest("Object -- resource.Quantity 2", &quantity2, &quantity1, []PathElement{NewPtrElem()}, quantity2, quantity1),
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

func buildSimpleTest(name string, base interface{}, update interface{}, path []PathElement, oldValue interface{}, newValue interface{}) diffTestObject {
	if reflect.TypeOf(base) != reflect.TypeOf(update) {
		panic(fmt.Errorf("base type (%T) != update type (%T)", base, update))
	}

	objectType := reflect.TypeOf(base)
	return diffTestObject{name: name, base: base, update: update,
		expect: ChangeSet{
			BaseType: objectType,
			Changes: []Change{
				NewValueChange(path, reflect.ValueOf(oldValue), reflect.ValueOf(newValue)),
			},
		},
	}
}
