package obj_diff

import (
	"fmt"
	. "github.com/takari/object-diff/pkg/obj_diff/helpers"
	"reflect"
)

// A ChangeSet represents the result of a diff as a set of Changes against a Base Type.
type ChangeSet struct {
	BaseType reflect.Type
	Changes  []Change
}

func (cs ChangeSet) String() string {
	return fmt.Sprintf("BaseType: %v Changes: %v", cs.BaseType, cs.Changes)
}

// Add a change to this change set.
func (cs *ChangeSet) AddPathChange(ctx []PathElement, oldValue reflect.Value, newValue reflect.Value) {
	cs.Changes = append(cs.Changes, NewValueChange(ctx, oldValue, newValue))
}

// Add an addition to this change set.
func (cs *ChangeSet) AddPathAddition(ctx []PathElement, newValue reflect.Value) {
	cs.Changes = append(cs.Changes, NewValueAddition(ctx, newValue))
}


// Add a delete to this change set.
func (cs *ChangeSet) AddPathDeletion(ctx []PathElement, oldValue reflect.Value) {
	cs.Changes = append(cs.Changes, NewValueDeletion(ctx, oldValue))
}

// Patch an object (in place/by reference) with the Changes within this
// ChangeSet. Panics if obj is not settable or does not match the BaseType.
func (cs ChangeSet) Patch(obj interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			e, ok := r.(PatchError)
			if ok {
				// fmt.Println("Recovered in Patch", r)
				err = e
			} else {
				panic(r)
			}
		}
	}()

	root := reflect.ValueOf(obj)
	if root.Kind() != reflect.Ptr || !root.Elem().CanSet() {
		return NewPatchError("can not set obj1 of Type: %v", root.Type())
	}

	if root.Type() != cs.BaseType {
		if root.Elem().Type() == cs.BaseType {
			root = root.Elem()
		} else {
			return NewPatchError("obj (%v) is not of type %v", reflect.TypeOf(obj), cs.BaseType)
		}
	}


	opConfig := ObjectPathConfig{true, true}

	for i := 0; i < len(cs.Changes); i++ {
		// fmt.Println()
		// fmt.Printf("##################\n")
		// fmt.Printf("### New Change ###\n")
		// fmt.Printf("##################\n")
		// fmt.Printf("Root: %+v\n", root)
		change := cs.Changes[i]
		// fmt.Printf("Change: %+v\n", change)
		op := NewObjectPathWithConfig(root, change.GetPath(), opConfig)
		// The first call to op.Next() skips past the pointer we were passed. If we
		// want to do anything with that pointer beforehand we must do it here.
		for op.Next() {
			// This loop is primarily ornamental, the call above to op.Next()
			// traverses the path, but there is nothing to do as the ObjectPath
			// takes care of everything. This is here primarily to be an extension
			// point for future changes.
			// fmt.Printf("Types lastVal: %T, currVal: %T\n", op.LastVal().Interface(), op.Interface())
			// fmt.Printf("Kinds lastVal: %v, currVal: %v\n", op.LastVal().Kind(), op.Kind())
			// fmt.Println("==================")

			switch op.Kind() {
			case reflect.Struct:
				// NO-OP
			case reflect.Map:
				// NO-OP
			case reflect.Array:
				// NO-OP
			case reflect.Slice:
				// NO-OP
			case reflect.Ptr:
				// NO-OP
			}
		}

		// Once we are at the end of the path we
		// either delete or update a value.
		if change.IsDeletion() {
			op.Delete()
		} else {
			op.Set(change.GetNewValue())
		}
	}

	return nil
}

// Compares this ChangeSet against another ChangeSet,
// returns true if the are the same. This is currently
// used only in the testing framework, but could have
// other uses.
func (cs ChangeSet) Equals(other ChangeSet) bool {
	if cs.BaseType !=  other.BaseType {
		return false
	}

	if len(cs.Changes) != len(other.Changes) {
		return false
	}

	for i := 0; i < len(cs.Changes); i++ {
		c1 := cs.Changes[i]
		c2 := other.Changes[i]

		if !c1.Equals(c2) {
			return false
		}
	}

	return true
}
