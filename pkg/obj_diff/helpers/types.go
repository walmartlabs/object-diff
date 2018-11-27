// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package helpers

import (
	"fmt"
	"reflect"
	"strings"
)

func NewValueChange(path []PathElement, oldValue reflect.Value, newValue reflect.Value) Change {
	var oldVal interface{}
	if oldValue.IsValid() {
		oldVal = oldValue.Interface()
	}

	var newVal interface{}
	if newValue.IsValid() {
		newVal = newValue.Interface()
	}

	return Change{path: path, oldValue: oldVal, newValue: newVal}
}

func NewValueAddition(path []PathElement, newValue reflect.Value) Change {
	var newVal interface{}
	if newValue.IsValid() {
		newVal = newValue.Interface()
	}

	return Change{path: path, newValue: newVal, addition: true}
}

func NewValueDeletion(path []PathElement, oldValue reflect.Value) Change {
	var oldVal interface{}
	if oldValue.IsValid() {
		oldVal = oldValue.Interface()
	}

	return Change{path: path, oldValue: oldVal, deletion: true}
}

// Change represents a single change to an object. It captures the
// path and either a newValue or a flag to delete the destination.
type Change struct {
	path     []PathElement
	oldValue interface{}
	newValue interface{}
	deletion bool
	addition bool
}

func (c Change) GetPath() []PathElement {
	return c.path
}

func (c Change) GetNewValue() reflect.Value {
	return reflect.ValueOf(c.newValue)
}

func (c Change) IsDeletion() bool {
	return c.deletion
}

// Compare this Change against another Change. Returns true if they
// are the same. Currently only used in testing.
func (c Change) Equals(other Change) bool {
	if c.deletion != other.deletion {
		return false
	}

	if !reflect.DeepEqual(c.newValue, other.newValue) {
		return false
	}

	if len(c.path) != len(other.path) {
		return false
	}

	for i := 0; i < len(c.path); i++ {
		pe1 := c.path[i]
		pe2 := other.path[i]

		if !pe1.Equals(pe2) {
			return false
		}
	}

	return true
}

// Return the path as a "human readable" string.
func (c Change) PathString() string {
	vsm := make([]string, len(c.path))
	for i, v := range c.path {
		vsm[i] = v.String()
	}
	return strings.Join(vsm, "")
}

func (c Change) String() string {
	if c.deletion {
		return fmt.Sprintf("%v -> [Deleted]", c.PathString())
	}

	return fmt.Sprintf("%v -> %v", c.PathString(), c.newValue)
}

// Create an Array/Slice Index PathElement.
func NewIndexElem(index int) PathElement {
	return PathElement{index: index}
}

// Create a Struct Field PathElement.
func NewFieldElem(index int, name string) PathElement {
	return PathElement{index: index, name: name}
}

// Create a Map Key PathElement.
func NewKeyElem(key interface{}) PathElement {
	keyVal, isValue := key.(reflect.Value)
	if isValue {
		if keyVal.IsValid() {
			key = keyVal.Interface()
		} else {
			key = nil
		}
	}
	return PathElement{index: -1, key: key}
}

// Create a Pointer PathElement.
func NewPtrElem() PathElement {
	return PathElement{index: -1, pointer: true}
}

// A PathElement represent a single step
// in a path through an object.
type PathElement struct {
	index   int
	name    string
	key     interface{}
	pointer bool
}

func (pe PathElement) GetIndex() int {
	return pe.index
}

func (pe PathElement) GetKey() reflect.Value {
	return reflect.ValueOf(pe.key)
}

func (pe PathElement) GetName() string {
	return pe.name
}

func (pe PathElement) IsPointer() bool {
	return pe.pointer
}

// Compares this PathElement against another PathElement. Returns true if
// they are the same. Currently only used in testing.
func (pe PathElement) Equals(other PathElement) bool {
	if pe.index != other.index {
		return false
	}

	if !reflect.DeepEqual(pe.key, other.key) {
		return false
	}

	if pe.pointer != other.pointer {
		return false
	}

	return pe.name == other.name
}

func (pe PathElement) String() string {
	if pe.key != nil {
		return fmt.Sprintf("{%v}", pe.key)
	}

	if pe.pointer {
		return "*"
	}

	if len(pe.name) > 0 {
		return fmt.Sprintf(".%v(%v)", pe.name, pe.index)
	}

	return fmt.Sprintf("[%v]", pe.index)
}

// Create a PathError object for capturing internal errors.
func NewPatchError(format string, args ...interface{}) PatchError {
	return PatchError{errStr: fmt.Sprintf(format, args...)}
}

var _ error = &PatchError{}

type PatchError struct {
	errStr string
}

func (err PatchError) Error() string {
	return err.errStr
}
