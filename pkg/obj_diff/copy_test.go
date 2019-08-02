// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"reflect"
	"testing"
)

type simpleStruct struct {
	A int64
	B float64
	C string
	D bool
	E resource.Quantity
}

type complexStruct struct {
	A simpleStruct
	B map[string]simpleStruct
	C [3]simpleStruct
	D []simpleStruct
	E *simpleStruct
	F *complexStruct
}

func TestCopyValueReflectively(t *testing.T) {
	ts1 := simpleStruct{123, 3.14, "ABC", true, resource.MustParse("500Mi")}
	ts2 := simpleStruct{456, 2.71, "DEF", false, resource.MustParse("1.5Gi")}
	ts3 := simpleStruct{042, 0.01, "XXX", false, resource.MustParse("3Ti")}
	map1 := map[string]int32{"a": 1, "b": 2, "c": 3}
	slice1 := []float32{3.14, 2.71, 0.01}
	array1 := [3]string{"ABC", "DEF", "XXX"}
	ptr1 := &ts1

	int1 := int32(1)
	structInt := structInt32{A: 1}
	mapInt := map[string]int32{"A": 1}
	arrayInt := [3]int32{1, 2, 3}
	sliceInt := []int32{1, 2, 3}
	ptrInt1 := &int1

	structStruct1 := structStruct{A: structInt}
	structMap1 := structMap{A: mapInt}
	structArray1 := structArray3{A: arrayInt}
	structSlice1 := structSlice{A: sliceInt}
	structPtr1 := structPtr{A: ptrInt1}

	mapStruct1 := map[string]structInt32{"A": structInt}
	mapMap1 := map[string]map[string]int32{"A": mapInt}
	mapArray1 := map[string][3]int32{"A": arrayInt}
	mapSlice1 := map[string][]int32{"A": sliceInt}
	mapPtr1 := map[string]*int32{"A": ptrInt1}

	arrayStruct1 := [3]structInt32{structInt, structInt, structInt}
	arrayMap1 := [3]map[string]int32{mapInt, mapInt, mapInt}
	arrayArray1 := [3][3]int32{arrayInt, arrayInt, arrayInt}
	arraySlice1 := [3][]int32{sliceInt, sliceInt, sliceInt}
	arrayPtr1 := [3]*int32{ptrInt1, ptrInt1, ptrInt1}

	sliceStruct1 := []structInt32{structInt, structInt, structInt}
	sliceMap1 := []map[string]int32{mapInt, mapInt, mapInt}
	sliceArray1 := [][3]int32{arrayInt, arrayInt, arrayInt}
	sliceSlice1 := [][]int32{sliceInt, sliceInt, sliceInt}
	slicePtr1 := []*int32{ptrInt1, ptrInt1, ptrInt1}

	ptrStruct1 := &structInt
	ptrMap1 := &mapInt
	ptrArray1 := &arrayInt
	ptrSlice1 := &sliceInt
	ptrPtr1 := &ptrInt1

	complex1 := complexStruct{
		ts1,
		map[string]simpleStruct{"a": ts2, "b": ts3},
		[3]simpleStruct{ts1, ts2, ts3},
		[]simpleStruct{ts1, ts2, ts3, ts1},
		&ts3,
		nil}
	complex2 := complexStruct{
		ts2,
		map[string]simpleStruct{"c": ts3, "d": ts1},
		[3]simpleStruct{ts2, ts3, ts1},
		[]simpleStruct{ts2, ts1, ts3, ts2},
		&ts2,
		&complex1}
	complex3 := complexStruct{
		ts3,
		map[string]simpleStruct{"e": ts1, "f": ts2},
		[3]simpleStruct{ts3, ts1, ts2},
		[]simpleStruct{ts2, ts2, ts2, ts2},
		&ts1,
		&complex2}

	tests := []struct {
		name   string
		object interface{}
	}{
		{name: "Basic -- Int", object: int(-123)},
		{name: "Basic -- Uint", object: uint(123)},
		{name: "Basic -- Float", object: float64(3.141586)},
		{name: "Basic -- Complex", object: complex128(123.45 + 456.78i)},
		{name: "Basic -- Bool", object: true},
		{name: "Basic -- String", object: "Hello World!"},

		{name: "Object -- Struct", object: ts1},
		{name: "Object -- Map", object: map1},
		{name: "Object -- Array", object: array1},
		{name: "Object -- Slice", object: slice1},
		{name: "Object -- Ptr", object: ptr1},

		{name: "Pairs -- Struct, Struct", object: structStruct1},
		{name: "Pairs -- Struct, Map", object: structMap1},
		{name: "Pairs -- Struct, Array", object: structArray1},
		{name: "Pairs -- Struct, Slice", object: structSlice1},
		{name: "Pairs -- Struct, Ptr", object: structPtr1},

		{name: "Pairs -- Map, Struct", object: mapStruct1},
		{name: "Pairs -- Map, Map", object: mapMap1},
		{name: "Pairs -- Map, Array", object: mapArray1},
		{name: "Pairs -- Map, Slice", object: mapSlice1},
		{name: "Pairs -- Map, Ptr", object: mapPtr1},

		{name: "Pairs -- Array, Struct", object: arrayStruct1},
		{name: "Pairs -- Array, Map", object: arrayMap1},
		{name: "Pairs -- Array, Array", object: arrayArray1},
		{name: "Pairs -- Array, Slice", object: arraySlice1},
		{name: "Pairs -- Array, Ptr", object: arrayPtr1},

		{name: "Pairs -- Slice, Struct", object: sliceStruct1},
		{name: "Pairs -- Slice, Map", object: sliceMap1},
		{name: "Pairs -- Slice, Array", object: sliceArray1},
		{name: "Pairs -- Slice, Slice", object: sliceSlice1},
		{name: "Pairs -- Slice, Ptr", object: slicePtr1},

		{name: "Pairs -- Ptr, Struct", object: ptrStruct1},
		{name: "Pairs -- Ptr, Map", object: ptrMap1},
		{name: "Pairs -- Ptr, Array", object: ptrArray1},
		{name: "Pairs -- Ptr, Slice", object: ptrSlice1},
		{name: "Pairs -- Ptr, Ptr", object: ptrPtr1},

		{name: "Quantity -- 500Mi", object: resource.MustParse("500Mi")},
		{name: "Quantity -- 1.5Gi", object: resource.MustParse("1.5Gi")},

		{name: "Complex -- 1 Level", object: complex1},
		{name: "Complex -- 2 Levels", object: complex2},
		{name: "Complex -- 3 Levels", object: &complex3},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			expect := test.object
			actual := CopyValueReflectively(expect)
			if !reflect.DeepEqual(expect, actual) {
				t.Logf("Expect: %+v", expect)
				t.Logf("Actual: %+v", actual)
				t.Fail()
			}
		})
	}
}

type aliasType1 int64
type aliasType2 simpleStruct

func TestTypeAlias(t *testing.T) {

	expect1 := aliasType1(123)
	actual1, ok := CopyValueReflectively(expect1).(aliasType1)
	if !ok {
		t.Fatal("Could not convert basic alias type!")
	}

	if !reflect.DeepEqual(expect1, actual1) {
		t.Logf("Expect: %+v", expect1)
		t.Logf("Actual: %+v", actual1)
		t.Fail()
	}

	expect2 := aliasType2{123, 3.14, "ABC", true, resource.MustParse("1.5Gi")}
	actual2, ok := CopyValueReflectively(expect2).(aliasType2)
	if !ok {
		t.Fatal("Could not convert object alias type!")
	}

	if !reflect.DeepEqual(expect2, actual2) {
		t.Logf("Expect: %+v", expect2)
		t.Logf("Actual: %+v", actual2)
		t.Fail()
	}

}
