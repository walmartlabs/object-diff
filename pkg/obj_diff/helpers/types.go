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

type Change interface {
	GetPath() []PathElement
	GetOldValue() reflect.Value
	GetNewValue() reflect.Value
	IsAddition() bool
	IsDeletion() bool
	PathString() string
	Equals(Change) bool
	fmt.Stringer
}

type SettableChange interface {
	Change
	SetNewValue(reflect.Value) error
}

func NewValueChange(path []PathElement, oldValue reflect.Value, newValue reflect.Value) SettableChange {
	var oldVal interface{}
	if oldValue.IsValid() {
		oldVal = oldValue.Interface()
	}

	var newVal interface{}
	if newValue.IsValid() {
		newVal = newValue.Interface()
	}

	return &change{path: path, oldValue: oldVal, newValue: newVal}
}

func NewValueAddition(path []PathElement, newValue reflect.Value) SettableChange {
	var newVal interface{}
	if newValue.IsValid() {
		newVal = newValue.Interface()
	}

	return &change{path: path, newValue: newVal, addition: true}
}

func NewValueDeletion(path []PathElement, oldValue reflect.Value) SettableChange {
	var oldVal interface{}
	if oldValue.IsValid() {
		oldVal = oldValue.Interface()
	}

	return &change{path: path, oldValue: oldVal, deletion: true}
}

// change represents a single change to an object. It captures the
// path and either a newValue or a flag to delete the destination.
type change struct {
	path     []PathElement
	oldValue interface{}
	newValue interface{}
	deletion bool
	addition bool
}

var _ SettableChange = &change{}

func (c change) GetPath() []PathElement {
	return c.path
}

func (c change) GetOldValue() reflect.Value {
	return reflect.ValueOf(c.oldValue)
}

func (c change) GetNewValue() reflect.Value {
	return reflect.ValueOf(c.newValue)
}

func (c *change) SetNewValue(newValue reflect.Value) error {
	if c.addition && !newValue.Type().AssignableTo(reflect.TypeOf(c.newValue)) {
		return fmt.Errorf("type of newValue (%T) can not be assigned to %T", newValue.Interface(), c.newValue)
	} else if !newValue.Type().AssignableTo(reflect.TypeOf(c.oldValue)) {
		return fmt.Errorf("type of newValue (%T) can not be assigned to %T", newValue.Interface(), c.oldValue)
	}

	c.newValue = newValue.Interface()
	return nil
}

func (c change) IsAddition() bool {
	return c.addition
}

func (c change) IsDeletion() bool {
	return c.deletion
}

// Compare this change against another change. Returns true if they
// are the same. Currently only used in testing.
func (c change) Equals(that Change) bool {
	if c.IsDeletion() != that.IsDeletion() {
		return false
	}

	thisValue := c.GetNewValue().Interface()
	thatValue := that.GetNewValue().Interface()
	if !reflect.DeepEqual(thisValue, thatValue) {
		return false
	}

	thisPath := c.GetPath()
	thatPath := that.GetPath()
	if len(thisPath) != len(thatPath) {
		return false
	}

	for i := 0; i < len(thisPath); i++ {
		pe1 := thisPath[i]
		pe2 := thatPath[i]

		if !pe1.Equals(pe2) {
			return false
		}
	}

	return true
}

// Return the path as a "human readable" string.
func (c change) PathString() string {
	vsm := make([]string, len(c.path))
	for i, v := range c.path {
		vsm[i] = v.String()
	}
	return strings.Join(vsm, "")
}

func (c change) String() string {
	if c.deletion {
		return fmt.Sprintf("%v -> [Deleted]", c.PathString())
	}

	return fmt.Sprintf("%v -> %v", c.PathString(), c.newValue)
}

type PathElement interface {
	GetIndex() int
	GetKey() reflect.Value
	GetName() string
	IsPointer() bool
	Equals(PathElement) bool
	fmt.Stringer
}

// Create an Array/Slice Index PathElement.
func NewIndexElem(index int) PathElement {
	return pathElement{index: index}
}

// Create a Struct Field PathElement.
func NewFieldElem(index int, name string) PathElement {
	return pathElement{index: index, name: name}
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
	return pathElement{index: -1, key: key}
}

// Create a Pointer PathElement.
func NewPtrElem() PathElement {
	return pathElement{index: -1, pointer: true}
}

// A PathElement represent a single step
// in a path through an object.
type pathElement struct {
	index   int
	name    string
	key     interface{}
	pointer bool
}

func (pe pathElement) GetIndex() int {
	return pe.index
}

func (pe pathElement) GetKey() reflect.Value {
	return reflect.ValueOf(pe.key)
}

func (pe pathElement) GetName() string {
	return pe.name
}

func (pe pathElement) IsPointer() bool {
	return pe.pointer
}

// Compares this PathElement against another PathElement. Returns true if
// they are the same. Currently only used in testing.
func (pe pathElement) Equals(other PathElement) bool {
	if pe.GetIndex() != other.GetIndex() {
		return false
	}

	thisKey := pe.GetKey()
	thatKey := other.GetKey()
	if thisKey.IsValid() || thatKey.IsValid() {
		if thisKey.IsValid() && thatKey.IsValid() {
			if !reflect.DeepEqual(thisKey.Interface(), thatKey.Interface()) {
				return false
			}
		}
	} else

	if pe.IsPointer() != other.IsPointer() {
		return false
	}

	return pe.GetName() == other.GetName()
}

func (pe pathElement) String() string {
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
