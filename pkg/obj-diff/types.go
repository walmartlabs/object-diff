package obj_diff

import (
	"fmt"
	"reflect"
	"strings"
)

type Change struct {
	Path     []PathElement
	NewValue reflect.Value
	Deleted  bool
}

func (c Change) Equals(other Change) bool {
	if c.Deleted != other.Deleted {
		return false
	}

	if !reflectValuesAreEqual(c.NewValue, other.NewValue) {
		return false
	}

	if len(c.Path) != len(other.Path) {
		return false
	}

	for i := 0; i < len(c.Path); i++ {
		pe1 := c.Path[i]
		pe2 := other.Path[i]

		if !pe1.Equals(pe2) {
			return false
		}
	}

	return true
}

func (c Change) PathString() string {
	vsm := make([]string, len(c.Path))
	for i, v := range c.Path {
		vsm[i] = v.String()
	}
	return strings.Join(vsm, "")
}

func (c Change) String() string {
	if c.Deleted {
		return fmt.Sprintf("%v -> [Deleted]", c.PathString())
	}

	return fmt.Sprintf("%v -> %v", c.PathString(), c.NewValue.Interface())
}

func NewIndexElem(index int) PathElement {
	return PathElement{Index: index}
}

func NewNameElem(index int, name string) PathElement {
	return PathElement{Index: index, Name: name}
}

func NewKeyElem(key interface{}) PathElement {
	keyVal, ok := key.(reflect.Value)
	if !ok {
		keyVal = reflect.ValueOf(key)
	}
	return PathElement{Index: -1, Key: keyVal}
}

func NewPtrElem() PathElement {
	return PathElement{Index: -1, Pointer: true}
}

type PathElement struct {
	Index   int
	Key     reflect.Value
	Pointer bool
	Name    string
}

func (pe PathElement) Equals(other PathElement) bool {
	if pe.Index != other.Index {
		return false
	}

	if !reflectValuesAreEqual(pe.Key, other.Key) {
		return false
	}

	if pe.Pointer != other.Pointer {
		return false
	}

	return pe.Name == other.Name
}

func (pe PathElement) String() string {
	if pe.Key.IsValid() {
		return fmt.Sprintf("{%v}", pe.Key.String())
	}

	if pe.Pointer {
		return "*"
	}

	if len(pe.Name) > 0 {
		return fmt.Sprintf(".%v(%v)", pe.Name, pe.Index)
	}

	return fmt.Sprintf("[%v]", pe.Index)
}

func reflectValuesAreEqual(v1, v2 reflect.Value) bool {
	if v1.IsValid() != v2.IsValid() {
		return false
	}

	if !v1.IsValid() {
		return true
	}

	if v1.Type() != v2.Type() {
		return false
	}

	switch v1.Kind() {
	case reflect.Map:
		fallthrough
	case reflect.Slice:
		fallthrough
	case reflect.Ptr:
		if v1.IsNil() != v2.IsNil() {
			return false
		}
		if v1.IsNil() {
			return true
		}
	}

	return reflect.DeepEqual(v1.Interface(), v2.Interface())
}

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
