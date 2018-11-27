// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	. "github.com/takari/object-diff/pkg/obj_diff/helpers"
	"reflect"
)

type ObjectPathConfig struct {
	// If the path would traverse into an object (struct, map,
	// array, slice, pointer) that does not exist, create it.
	CreateMissingObjects bool
	// If the path traverses into an invalid Map key or Slice
	// index, create the object that should be there.
	CreateMissingValues bool
}

// var DEFAULT_CONFIG = ObjectPathConfig{false, false}

// func NewObjectPath(root reflect.Value, path []PathElement) *ObjectPath {
// 	return NewObjectPathWithConfig(root, path, DEFAULT_CONFIG)
// }

// Creates an Object path, if root is not writable then mutating actions on
// the resulting ObjectPath will panic. The path is a list of PathElements to
// follow when traversing root. Finally config contains options and actions
// that ObjectPath can take for you automatically.
func NewObjectPathWithConfig(root reflect.Value, path []PathElement, config ObjectPathConfig) *ObjectPath {
	objectPath := &ObjectPath{Value: root, lastVals: []reflect.Value{}, index: -1, Path: path, config: config}
	// We need to run this here because the first call to Next() will operate on the second value.
	objectPath.nextConfigOptions()
	return objectPath
}

// An Object path provides a way to step through a given []PathElement as it
// relates to object given at creation and take actions at each point along the
// way. At any given point the ObjectPath can be used as a reflect.Value at the
// current point in the traversal of path.
type ObjectPath struct {
	reflect.Value
	lastVals []reflect.Value
	index    int
	Path     []PathElement
	config   ObjectPathConfig
}

// Advance to the next object in the path. Returns true if there are further
// elements in the path, and false otherwise.
func (op *ObjectPath) Next() (hasNext bool) {
	op.lastVals = append(op.lastVals, op.Value)
	switch op.Kind() {
	case reflect.Struct:
		op.Value = op.GetField()
	case reflect.Map:
		op.Value = op.GetMapValue()
	case reflect.Array:
		op.Value = op.GetIndex()
	case reflect.Slice:
		op.Value = op.GetIndex()
	case reflect.Ptr:
		if !op.IsPointer() {
			panic(NewPatchError("unexpected pointer in Next()"))
		}
		op.Value = op.Elem()
	default:
		panic(NewPatchError("unhandled path kind '%v'\n", op.Kind()))
	}
	op.index++
	op.nextConfigOptions()
	return op.index+1 < len(op.Path)
}

// Apply optional config items after advancement.
func (op ObjectPath) nextConfigOptions() {
	switch op.Kind() {
	case reflect.Struct:

	case reflect.Map:
		if op.config.CreateMissingObjects {
			op.CreateIfMissing()
		}
		if op.config.CreateMissingValues && !op.GetMapValue().IsValid() {
			op.SetMapValueToNew(op.Type().Elem())
		}
	case reflect.Array:

	case reflect.Slice:
		if op.config.CreateMissingObjects {
			op.CreateIfMissing()
		}
		if op.config.CreateMissingValues && op.NeedsAppend() {
			op.AppendNew(op.Type().Elem())
		}
	case reflect.Ptr:
		if op.config.CreateMissingObjects {
			op.CreateIfMissing()
		}
	}
}

// Retrieve the current Path Element.
func (op ObjectPath) PathElem() PathElement {
	return op.Path[op.index+1]
}

// Retrieve the last value of this ObjectPath.
func (op ObjectPath) LastVal() reflect.Value {
	return op.lastVals[op.index]
}

// Retrieve the next from a Struct type. Panics if the
// current object is not a Struct.
func (op *ObjectPath) GetField() reflect.Value {
	return op.Field(op.PathElem().GetIndex())
}

// Retrieves the next index. Panics if the current object
// is not an Array or Slice.
func (op *ObjectPath) GetIndex() reflect.Value {
	return op.Index(op.PathElem().GetIndex())
}

// Retrieves the value for the next key. Panics if the
// current object is not a Map.
func (op *ObjectPath) GetMapValue() reflect.Value {
	return op.MapIndex(op.PathElem().GetKey())
}

// Sets the next map key to the value of a new reflect.Type.
// Panics if newType is not assignable to the Map value.
func (op *ObjectPath) SetMapValueToNew(newType reflect.Type) {
	op.SetMapValue(buildNewValue(newType))
}

// Sets the next map key to the given value. Panics if the
// newValue is not assignable to the Map value.
func (op *ObjectPath) SetMapValue(newValue reflect.Value) {
	op.SetMapIndex(op.PathElem().GetKey(), newValue)
}

// Returns true if the current PathElement is a pointer.
func (op *ObjectPath) IsPointer() bool {
	return op.PathElem().IsPointer()
}

// Returns true if the next index is at the end of a Slice.
// This means that an append will be successful at this point.
func (op *ObjectPath) NeedsAppend() bool {
	if op.Len() < op.PathElem().GetIndex() {
		panic(NewPatchError("index (%v) larger than slice size(%v)", op.PathElem().GetIndex(), op.Len()))
	}
	return op.Len() == op.PathElem().GetIndex()
}

// Appends to the current Slice a new object of newType.
// Panics if newType is not assignable to the Slice type.
func (op *ObjectPath) AppendNew(newType reflect.Type) {
	op.Append(buildNewValue(newType))
}

// Appends to the current Slice a newValue.
// Panics if newValue is not assignable to the Slice type.
func (op *ObjectPath) Append(newVal reflect.Value) {
	op.Set(reflect.Append(op.Value, newVal))
}

// Returns true if the next index exists in the current Slice
// or Array. Panics if current element is not a Slice or Array.
func (op *ObjectPath) InBounds() bool {
	return op.Len() > op.PathElem().GetIndex()
}

// Create the next object in the path if it does not currently exist.
func (op *ObjectPath) CreateIfMissing() {
	if op.IsNil() {
		op.SetToNew(op.Type())
	}
}

// Set the current value to an instance of newType. Panics
// if newType is not assignable to the current value.
func (op *ObjectPath) SetToNew(newType reflect.Type) {
	op.Set(buildNewValue(newType))
}

// Set the current value to newValue. Panics if newValue
// is not assignable to the current value.
func (op *ObjectPath) Set(newVal reflect.Value) {
	// fmt.Println("\n### In set() ###")
	// fmt.Printf("CURRENT: %T, settable: %v\n", op.Interface(), op.CanSet())
	// fmt.Printf("newVal: %+v\n", newVal)

	settable := op.Value
	prevVal := reflect.ValueOf(nil)
	// This loop primarily exists to backtrack to an object which is settable.
	// This most likely occurs when we're trying to set the value of a Map as
	// Map values are not directly settable. However this could occur in other
	// situations.
	for i := op.index; !settable.CanSet(); i-- {
		if i < 0 {
			panic("No settable object available!")
		}
		settable = op.lastVals[i]
		// As we backtrack it is necessary to recreate the objects we have passed
		// as they are not settable and thus copying/cloning them is the only option.
		prevVal = CopyReflectValue(op.lastVals[i])
		switch prevVal.Kind() {
		case reflect.Struct:
			prevVal.Field(op.Path[i].GetIndex()).Set(newVal)
		case reflect.Map:
			prevVal.SetMapIndex(op.Path[i].GetKey(), newVal)
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			prevVal.Index(op.Path[i].GetIndex()).Set(newVal)
		case reflect.Ptr:
			prevVal.Elem().Set(newVal)
		default:
			panic(NewPatchError("unhandled set-backtrack kind '%v'\n", prevVal.Kind()))
		}
		newVal = prevVal

		// fmt.Printf("CURRENT: %T, settable: %v\n", settable.Interface(), settable.CanSet())
		// fmt.Printf("newVal: %+v\n", newVal)
	}

	settable.Set(newVal)
	// fmt.Println("### Leaving set() ###")
}

// Delete the object at the current point in the path. Delete
// is only supported for Map, Slice, and Ptr; panics otherwise.
func (op *ObjectPath) Delete() {
	// fmt.Println("\n### In delete() ###")
	lastVal := op.LastVal()
	switch lastVal.Kind() {
	case reflect.Map:
		if lastVal.MapIndex(op.Path[op.index].GetKey()).IsValid() {
			// Setting a Map value to the 'nil' value clears the key.
			op.Set(reflect.Value{})
		}
	case reflect.Slice:
		lastVal.Set(lastVal.Slice(0, lastVal.Len()-1))
	case reflect.Ptr:
		lastVal.Set(reflect.Zero(lastVal.Type()))
	default:
		panic(NewPatchError("unhandled delete kind '%v'", lastVal.Kind()))
	}
	// fmt.Println("### Leaving delete() ###")
}

// Build a new value of type newType.
func buildNewValue(newType reflect.Type) (newValue reflect.Value) {
	// fmt.Printf("Building new %v\n", newType)
	switch newType.Kind() {
	case reflect.Map:
		newValue = reflect.MakeMap(newType)
	case reflect.Slice:
		newValue = reflect.MakeSlice(newType, 0, 0)
	case reflect.Ptr:
		newValue = reflect.New(newType.Elem())
	default:
		newValue = reflect.New(newType).Elem()
	}
	// fmt.Printf("Built %+v\n", newValue.Interface())
	return
}
