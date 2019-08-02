// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"testing"
)

type NestObj struct {
	Int int64
	Str string
}

type Obj struct {
	Int        int32
	IntPtr     *int16
	Float      float64
	Str        string
	Bool       bool
	IntList    []int64
	BoolList   [3]bool
	StrIntMap  map[string]int64
	NestedObj  NestObj
	NestedPtr1 *NestObj
	NestedPtr2 *NestObj
	NestedPtr3 *NestObj
	MapOfMaps  map[string]map[string]NestObj
	Quantity   resource.Quantity
}

func TestDiffThenPatch(t *testing.T) {
	four := int16(4)
	five := int16(5)
	nest1 := NestObj{9, "A"}
	nest2 := NestObj{9, "B"}
	nest3 := NestObj{8, "C"}
	o1 := Obj{1, &four, 1.2, "Foo", false,
		[]int64{1, 2}, [3]bool{true, false, true},
		map[string]int64{"a": 1, "b": 2, "d": 4},
		NestObj{3, "Hello"}, &nest1, nil, &nest2,
		map[string]map[string]NestObj{"a": {"b": nest1}, "c": {"d": nest2}}, resource.MustParse("500Mi")}
	o2 := Obj{2, &five, 3.14, "Bar", true,
		[]int64{3, 4, 5}, [3]bool{true, true, false},
		map[string]int64{"a": 2, "c": 3, "d": 4},
		NestObj{7, "World"}, &nest2, &nest1, nil,
		map[string]map[string]NestObj{"a": {"b": nest3}, "c": {"d": nest2}}, resource.MustParse("1.5Gi")}

	diff, err := Diff(o1, o2)
	if err != nil {
		t.Fatalf("Error in Diff: %v", err)
	}

	fmt.Printf("%v+\n", diff)

	t.Logf("BaseType: %v", diff.BaseType)
	t.Logf("Changes:")
	for _, change := range diff.Changes {
		t.Logf("%+v", change)
	}

	o3 := Obj{}
	err = diff.Patch(&o3)
	t.Logf("Original: %+v", o1)
	t.Logf("Expected: %+v", o2)
	t.Logf("Applied: %+v", o3)
	if err != nil {
		t.Fatalf("Error in Patch: %v", err)
	}

}

func TestDiffPointersThenPatch(t *testing.T) {
	four := int16(4)
	five := int16(5)
	nest1 := NestObj{9, "A"}
	nest2 := NestObj{9, "B"}
	nest3 := NestObj{8, "C"}
	o1 := Obj{1, &four, 1.2, "Foo", false,
		[]int64{1, 2}, [3]bool{true, false, true},
		map[string]int64{"a": 1, "b": 2, "d": 4},
		NestObj{3, "Hello"}, &nest1, nil, &nest2,
		map[string]map[string]NestObj{"a": {"b": nest1}, "c": {"d": nest2}}, resource.MustParse("1.5Gi")}
	o2 := Obj{2, &five, 3.14, "Bar", true,
		[]int64{3, 4, 5}, [3]bool{true, true, false},
		map[string]int64{"a": 2, "c": 3, "d": 4},
		NestObj{7, "World"}, &nest2, &nest1, nil,
		map[string]map[string]NestObj{"a": {"b": nest3}, "c": {"d": nest2}}, resource.MustParse("500Mi")}

	diff, err := Diff(&o1, &o2)
	if err != nil {
		t.Fatalf("Error in Diff: %v", err)
	}

	fmt.Printf("%v+\n", diff)

	t.Logf("BaseType: %v", diff.BaseType)
	t.Logf("Changes:")
	for _, change := range diff.Changes {
		t.Logf("%+v", change)
	}

	o3 := Obj{}
	err = diff.Patch(&o3)
	t.Logf("Original: %+v", o1)
	t.Logf("Expected: %+v", o2)
	t.Logf("Applied: %+v", o3)
	if err != nil {
		t.Fatalf("Error in Patch: %v", err)
	}

}

func TestAddByteArrayToMap(t *testing.T) {
	m1 := map[string][]byte{}
	m2 := map[string][]byte{
		"foo": []byte("test"),
	}

	diff, err := Diff(&m1, &m2)
	if err != nil {
		t.Fatalf("Error in Diff: %v", err)
	}

	fmt.Printf("%v+\n", diff)

	t.Logf("BaseType: %v", diff.BaseType)
	t.Logf("Changes:")
	for _, change := range diff.Changes {
		t.Logf("%+v", change)
	}

	m3 := map[string][]byte{
		"foo": []byte("test"),
	}

	err = diff.Patch(&m3)
	t.Logf("Original: %+v", m1)
	t.Logf("Expected: %+v", m2)
	t.Logf("Applied: %+v", m3)
	if err != nil {
		t.Fatalf("Error in Patch: %v", err)
	}

	m4 := map[string][]byte{}
	err = diff.Patch(&m4)
	t.Logf("Original: %+v", m1)
	t.Logf("Expected: %+v", m2)
	t.Logf("Applied: %+v", m4)
	if err != nil {
		t.Fatalf("Error in Patch: %v", err)
	}
}
