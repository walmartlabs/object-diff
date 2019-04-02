// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	types := makeTypes(2, false)
	for _, typeDef := range types {
		origin := makeObject(makeType(typeDef, false), 0)
		change := makeObject(makeType(typeDef, false), 0)
		fmt.Printf("Def: %v, Object: %+v\n", typeDef, origin)
		fmt.Printf("Def: %v, Object: %+v\n", typeDef, change)

	}
	fmt.Printf("Total Count %v\n", len(types))
}

// buildSimpleTest("Pairs -- Struct, Struct", structStruct{A: structInt32{A: int1}}, structStruct{A: structInt32{A: int2}}, []PathElement{NewFieldElem(0, "A"), NewFieldElem(0, "A")}, int2),
// buildSimpleTest("Pairs -- Struct, Map", structMap{map[string]int32{"A": int1}}, structMap{map[string]int32{"A": int2}}, []PathElement{NewFieldElem(0, "A"), NewKeyElem("A")}, int2),
// buildSimpleTest("Pairs -- Struct, Array", structArray1{[1]int32{int1}}, structArray1{[1]int32{int2}}, []PathElement{NewFieldElem(0, "A"), NewIndexElem(0)}, int2),
// buildSimpleTest("Pairs -- Struct, Slice", structSlice{[]int32{int1}}, structSlice{[]int32{int2}}, []PathElement{NewFieldElem(0, "A"), NewIndexElem(0)}, int2),
// buildSimpleTest("Pairs -- Struct, Ptr", structPtr{&int1}, structPtr{&int2}, []PathElement{NewFieldElem(0, "A"), NewPtrElem()}, int2),
//
// buildSimpleTest("Pairs -- Map, Struct", structStruct{A: structInt32{A: int1}}, structStruct{A: structInt32{A: int2}}, []PathElement{NewFieldElem(0, "A"), NewFieldElem(0, "A")}, int2),
// buildSimpleTest("Pairs -- Map, Map", structMap{map[string]int32{"A": int1}}, structMap{map[string]int32{"A": int2}}, []PathElement{NewFieldElem(0, "A"), NewKeyElem("A")}, int2),
// buildSimpleTest("Pairs -- Map, Array", structArray1{[1]int32{int1}}, structArray1{[1]int32{int2}}, []PathElement{NewFieldElem(0, "A"), NewIndexElem(0)}, int2),
// buildSimpleTest("Pairs -- Map, Slice", structSlice{[]int32{int1}}, structSlice{[]int32{int2}}, []PathElement{NewFieldElem(0, "A"), NewIndexElem(0)}, int2),
// buildSimpleTest("Pairs -- Map, Ptr", structPtr{&int1}, structPtr{&int2}, []PathElement{NewFieldElem(0, "A"), NewPtrElem()}, int2),

var BasicTypes = []reflect.Kind{reflect.Int, reflect.Uint, reflect.Float64, reflect.Complex128, reflect.Bool, reflect.String}
var ObjectTypes = []reflect.Kind{reflect.Struct, reflect.Map, reflect.Array, reflect.Slice, reflect.Ptr}
var ComparableTypes = []reflect.Kind{reflect.Struct, reflect.Array}

func buildTestCombinations(depth uint, structFieldCnt int32) []diffTestObject {
	doBuildTestCombinations(depth, structFieldCnt)

	return []diffTestObject{}
}
func doBuildTestCombinations(depth uint, structFieldCnt int32) (bases []interface{}) {
	return nil
}

type typeField struct {
	Name string
	Elem typeLevel
}

type typeLevel struct {
	Kind   reflect.Kind
	Key    *typeLevel
	Elem   *typeLevel
	Fields []typeField
	Size   int
}

func (tl typeLevel) String() string {
	bldr := strings.Builder{}
	bldr.WriteString(tl.Kind.String())
	switch tl.Kind {
	case reflect.Struct:
		bldr.WriteString("{")
		first := true
		for _, field := range tl.Fields {
			if first {
				first = false
			} else {
				bldr.WriteString(", ")
			}
			bldr.WriteString(fmt.Sprintf("%v: %v", field.Name, field.Elem.String()))
		}
		bldr.WriteString("}")
	case reflect.Map:
		bldr.WriteString(fmt.Sprintf("{%v -> %v}", tl.Key.String(), tl.Elem.String()))
	case reflect.Array:
		bldr.WriteString(fmt.Sprintf("[%v]{%v}", tl.Size, tl.Elem.String()))
	case reflect.Slice:
		bldr.WriteString(fmt.Sprintf("[]{%v}", tl.Elem.String()))
	case reflect.Ptr:
		bldr.WriteString(fmt.Sprintf("*{%v}", tl.Elem.String()))
	}

	return bldr.String()
}

var whichBasicType = 0

func makeTypes(depth int, comparable bool) []typeLevel {
	// typeDef := typeLevel{
	// 	Kind: reflect.Struct,
	// 	Fields: []typeField{
	// 		{Name: "A", Elem: typeLevel{Kind: reflect.Int32}},
	// 		{Name: "B", Elem: typeLevel{
	// 			Kind: reflect.Map,
	// 			key: &typeLevel{Kind: reflect.Struct, Fields: []typeField{
	// 				{Name: "A", Elem: typeLevel{Kind: reflect.Int32}},
	// 			}},
	// 			Elem: &typeLevel{Kind: reflect.String}}}}}
	// fmt.Printf("======== Depth: %v ========\n", depth)
	var typeDefs []typeLevel
	// if depth == 0 {
	// 	typeDefs = append(typeDefs, typeLevel{Kind: BasicTypes[whichBasicType]})
		// whichBasicType = (whichBasicType + 1) % len(BasicTypes)

	var basicTypes []reflect.Kind
	var outerTypes []reflect.Kind
	var innerTypes []typeLevel
	if depth <= 0 {
		basicTypes = BasicTypes[whichBasicType:whichBasicType+1]
	} else {
		if comparable {
			innerTypes = makeTypes(depth - 1, true)
			outerTypes = ComparableTypes
			basicTypes = BasicTypes[whichBasicType:whichBasicType+1]
		} else {
			innerTypes = makeTypes(depth-1, false)
			outerTypes = ObjectTypes
		}
	}
	for _, outerType := range outerTypes {
		for _, innerType := range innerTypes {
			// fmt.Printf("outer: %v, inner: %v\n", outerType, innerType)
			switch outerType {
			case reflect.Struct:
				typeDefs = append(typeDefs, typeLevel{Kind: reflect.Struct, Fields: []typeField{{Name: "A", Elem: innerType}}})
			case reflect.Map:
				keyTypes := makeTypes(depth-1, true)
				for _, keyType := range keyTypes {
					keyCopy := CopyValueReflectively(keyType).(typeLevel)
					elemCopy := CopyValueReflectively(innerType).(typeLevel)
					typeDefs = append(typeDefs, typeLevel{Kind: reflect.Map, Key: &keyCopy, Elem: &elemCopy})
				}
			case reflect.Array:
				innerCopy := CopyValueReflectively(innerType).(typeLevel)
				typeDefs = append(typeDefs, typeLevel{Kind: reflect.Array, Elem: &innerCopy, Size: 3})
			case reflect.Slice:
				innerCopy := CopyValueReflectively(innerType).(typeLevel)
				typeDefs = append(typeDefs, typeLevel{Kind: reflect.Slice, Elem: &innerCopy})
			case reflect.Ptr:
				innerCopy := CopyValueReflectively(innerType).(typeLevel)
				typeDefs = append(typeDefs, typeLevel{Kind: reflect.Ptr, Elem: &innerCopy})
			default:
				panic(fmt.Errorf("unexpected object type %v", outerType))
			}
		}
	}

	for _, basicType := range basicTypes {
		typeDefs = append(typeDefs, typeLevel{Kind: basicType})
	}

	return typeDefs
}

func makeType(typeToMake typeLevel, mustBeComparable bool) reflect.Type {
	switch typeToMake.Kind {
	case reflect.Struct:
		var structFields []reflect.StructField
		for _, field := range typeToMake.Fields {
			structFields = append(structFields, reflect.StructField{Name: field.Name, Type: makeType(field.Elem, mustBeComparable)})
		}
		return reflect.StructOf(structFields)
	case reflect.Map:
		if mustBeComparable {
			panic(fmt.Errorf("map is not comparable"))
		}
		return reflect.MapOf(makeType(*typeToMake.Key, true), makeType(*typeToMake.Elem, false))
	case reflect.Array:
		return reflect.ArrayOf(typeToMake.Size, makeType(*typeToMake.Elem, mustBeComparable))
	case reflect.Slice:
		if mustBeComparable {
			panic(fmt.Errorf("slice is not comparable"))
		}
		return reflect.SliceOf(makeType(*typeToMake.Elem, false))
	case reflect.Ptr:
		return reflect.PtrTo(makeType(*typeToMake.Elem, mustBeComparable))
	default:
		return makeBasicType(typeToMake)
	}
	return nil
}

const ObjectMapEntries = 3
const ObjectSliceEntries = 3

func makeObject(newType reflect.Type, value int) (obj reflect.Value) {
	switch newType.Kind() {
	case reflect.Struct:
		obj = reflect.New(newType).Elem()
		for f := 0; f < newType.NumField(); f++ {
			obj.Field(f).Set(makeObject(newType.Field(f).Type, value+f))
		}
	case reflect.Map:
		obj = reflect.MakeMap(newType)
		for i := 0; i < ObjectMapEntries; i++ {
			key := makeObject(newType.Key(), i)
			value := makeObject(newType.Elem(), value+i)
			obj.SetMapIndex(key, value)
		}
	case reflect.Array:
		obj = reflect.New(newType).Elem()
		for i := 0; i < newType.Len(); i++ {
			obj.Index(i).Set(makeObject(newType.Elem(), value+i))
		}
	case reflect.Slice:
		obj = reflect.MakeSlice(newType, 0, 0)
		for i := 0; i < ObjectSliceEntries; i++ {
			obj = reflect.Append(obj, makeObject(newType.Elem(), value+i))
		}
	case reflect.Ptr:
		obj = reflect.New(newType.Elem())
		obj.Elem().Set(makeObject(newType.Elem(), value))
	default:
		obj = makeBasicValue(newType.Kind(), value)
	}

	return
}

func makeBasicType(typeToMake typeLevel) reflect.Type {
	return makeBasicValue(typeToMake.Kind, 0).Type()
}

func makeBasicValue(kindToMake reflect.Kind, value int) reflect.Value {
	switch kindToMake {
	case reflect.String:
		return reflect.ValueOf(string('A' + value))
	case reflect.Int64:
		return reflect.ValueOf(int64(value))
	case reflect.Int32:
		return reflect.ValueOf(int32(value))
	case reflect.Int16:
		return reflect.ValueOf(int16(value))
	case reflect.Int8:
		return reflect.ValueOf(int8(value))
	case reflect.Int:
		return reflect.ValueOf(int(value))

	case reflect.Uint64:
		return reflect.ValueOf(uint64(value))
	case reflect.Uint32:
		return reflect.ValueOf(uint32(value))
	case reflect.Uint16:
		return reflect.ValueOf(uint16(value))
	case reflect.Uint8:
		return reflect.ValueOf(uint8(value))
	case reflect.Uint:
		return reflect.ValueOf(uint(value))

	case reflect.Float64:
		return reflect.ValueOf(float64(value))
	case reflect.Float32:
		return reflect.ValueOf(float32(value))

	case reflect.Complex128:
		return reflect.ValueOf(complex128(complex(float64(value), float64(value))))
	case reflect.Complex64:
		return reflect.ValueOf(complex64(complex(float32(value), float32(value))))

	case reflect.Bool:
		return reflect.ValueOf(value%2 == 1)

	default:
		panic(fmt.Errorf("unhandled kind '%v'\n", kindToMake))
	}
}
