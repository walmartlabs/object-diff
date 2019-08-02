// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package obj_diff

import (
	"fmt"
	. "github.com/walmartlabs/object-diff/pkg/obj_diff/helpers"
	"reflect"
)

// BUG(11xor6) Renaming of Map keys results in a deletion and addition.
// BUG(11xor6) Lists with different orders but the same elements will generate changes.

// Computes the change set between two objects, both objects must have the same type.
// This returns a ChangeSet on success and an error on failure.
func Diff(obj1 interface{}, obj2 interface{}) (*ChangeSet, error) {
	v1 := reflect.ValueOf(obj1)
	v2 := reflect.ValueOf(obj2)

	if v1.Type() != v2.Type() {
		return nil, fmt.Errorf("type of obj1(%T) not equal to obj2(%T)", obj1, obj2)
	}

	changeSet := &ChangeSet{BaseType: v1.Type()}
	return changeSet, doDiff(v1.Type(), v1, v2, changeSet, []PathElement{})
}


func doDiff(currType reflect.Type, v1 reflect.Value, v2 reflect.Value, cs *ChangeSet, ctx []PathElement) error {
	if !(v1.CanInterface() && v2.CanInterface()) {
		return InterfaceError{}
	}

	switch currType.Kind() {
	case reflect.Struct:
		for f := 0; f < currType.NumField(); f++ {
			currField := currType.Field(f)
			newCtx := extendContext(ctx, NewFieldElem(f, currField.Name))
			err := doDiff(currField.Type, v1.Field(f), v2.Field(f), cs, newCtx)
			if err != nil {
				if IsInterfaceError(err) {
					if !reflect.DeepEqual(v1.Interface(), v2.Interface()) {
						cs.AddPathChange(ctx, v1, v2)
					}
					break // We break because all fields of this obj are not Interface-able.
				} else {
					return err
				}
			}
		}
	case reflect.Map:
		for _, key := range v1.MapKeys() {
			val2 := v2.MapIndex(key)
			newCtx := extendContext(ctx, NewKeyElem(key))
			if val2.IsValid() {
				// Exists in both v1 and v2, do they match?
				err := doDiff(currType.Elem(), v1.MapIndex(key), v2.MapIndex(key), cs, newCtx)
				if err != nil {
					if IsInterfaceError(err) {
						// Only structs should create interface errors
						panic(err)
					} else {
						return err
					}
				}
			} else {
				// Exists in v1 and not in v2.
				cs.AddPathDeletion(newCtx, v1.MapIndex(key))
			}
		}

		for _, key := range v2.MapKeys() {
			val1 := v1.MapIndex(key)
			if !val1.IsValid() {
				// Exists in v2 and not in v1.
				newCtx := extendContext(ctx, NewKeyElem(key))
				cs.AddPathAddition(newCtx, v2.MapIndex(key))
			}
		}
	case reflect.Array:
		for i := 0; i < currType.Len(); i++ {
			newCtx := extendContext(ctx, NewIndexElem(i))
			err := doDiff(currType.Elem(), v1.Index(i), v2.Index(i), cs, newCtx)
			if err != nil {
				if IsInterfaceError(err) {
					// Only structs should create interface errors
					panic(err)
				} else {
					return err
				}
			}
		}
	case reflect.Slice:
		minLen := intMin(v1.Len(), v2.Len())
		maxLen := intMax(v1.Len(), v2.Len())
		for i := 0; i < minLen; i++ {
			newCtx := extendContext(ctx, NewIndexElem(i))
			err := doDiff(currType.Elem(), v1.Index(i), v2.Index(i), cs, newCtx)
			if err != nil {
				if IsInterfaceError(err) {
					// Only structs should create interface errors
					panic(err)
				} else {
					return err
				}
			}
		}

		if minLen != maxLen {
			if maxLen == v1.Len() {
				for i := minLen; i < maxLen; i++ {
					newCtx := extendContext(ctx, NewIndexElem(i))
					cs.AddPathDeletion(newCtx, v1.Index(i))
				}
			} else { // maxLen == v2.Len()
				for i := minLen; i < maxLen; i++ {
					newCtx := extendContext(ctx, NewIndexElem(i))
					cs.AddPathAddition(newCtx, v2.Index(i))
				}

			}
		}
	case reflect.Ptr:
		newCtx := extendContext(ctx, NewPtrElem())
		if v1.IsNil() && v2.IsNil() {
			return nil
		} else if v1.IsNil() {
			cs.AddPathAddition(newCtx, v2.Elem())
		} else if v2.IsNil() {
			cs.AddPathDeletion(newCtx, v1.Elem())
		} else {
			err := doDiff(currType.Elem(), v1.Elem(), v2.Elem(), cs, newCtx)
			if err != nil {
				if IsInterfaceError(err) {
					// Only structs should create interface errors
					panic(err)
				} else {
					return err
				}
			}
		}
	default:
		return compareBasicType(currType, v1, v2, cs, ctx)
	}

	return nil
}

// This creates a copy of the context and adds the new element to it. It is
// important to make a copy as the same context could be used by multiple
// changes and could modify each other.
func extendContext(ctx []PathElement, pe PathElement) []PathElement {
	newCtx := make([]PathElement, len(ctx), len(ctx)+1)
	copy(newCtx, ctx)
	return append(newCtx, pe)
}

func intMin(x int, y int) int {
	if x < y {
		return x
	}

	return y
}
func intMax(x int, y int) int {
	if x > y {
		return x
	}

	return y
}

func compareBasicType(currType reflect.Type, v1 reflect.Value, v2 reflect.Value, cs *ChangeSet, ctx []PathElement) error {
	switch currType.Kind() {
	case reflect.String:
		if v1.String() != v2.String() {
			cs.AddPathChange(ctx, v1, v2)
		}
	case reflect.Int64:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int:
		if v1.Int() != v2.Int() {
			cs.AddPathChange(ctx, v1, v2)
		}

	case reflect.Uint64:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint16:
		fallthrough
	case reflect.Uint8:
		fallthrough
	case reflect.Uint:
		if v1.Uint() != v2.Uint() {
			cs.AddPathChange(ctx, v1, v2)
		}

	case reflect.Float64:
		fallthrough
	case reflect.Float32:
		if v1.Float() != v2.Float() {
			cs.AddPathChange(ctx, v1, v2)
		}

	case reflect.Complex128:
		fallthrough
	case reflect.Complex64:
		if v1.Complex() != v2.Complex() {
			cs.AddPathChange(ctx, v1, v2)
		}

	case reflect.Bool:
		if v1.Bool() != v2.Bool() {
			cs.AddPathChange(ctx, v1, v2)
		}

	default:
		return fmt.Errorf("unhandled kind '%v'\n", currType.Kind())
	}

	return nil
}
