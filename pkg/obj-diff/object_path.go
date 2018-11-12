package obj_diff

import (
	"fmt"
	"reflect"
)

type ObjectPathConfig struct {
	CreateMissingObjects bool
	CreateMissingValues bool
}

// var DEFAULT_CONFIG = ObjectPathConfig{false, false}

// func NewObjectPath(root reflect.Value, path []PathElement) *objectPath {
// 	return NewObjectPathWithConfig(root, path, DEFAULT_CONFIG)
// }

func NewObjectPathWithConfig(root reflect.Value, path []PathElement, config ObjectPathConfig) *objectPath {
	pathWithPtr := append([]PathElement{{Pointer: true}}, path...)
	return &objectPath{Value: root, lastVals: []reflect.Value{}, index: -1, Path: pathWithPtr, config: config}
}

type objectPath struct {
	reflect.Value
	lastVals []reflect.Value
	index    int
	Path     []PathElement
	config   ObjectPathConfig
}

func (op *objectPath) Next() (hasNext bool) {
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

func (op objectPath) nextConfigOptions() {
	switch op.Kind() {
	case reflect.Struct:

	case reflect.Map:
		if op.config.CreateMissingObjects {
			op.CreateIfMissing()
		}
		if op.config.CreateMissingValues && !op.GetMapValue().IsValid(){
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

func (op objectPath) PathElem() PathElement {
	return op.Path[op.index+1]
}

func (op objectPath) LastVal() reflect.Value {
	return op.lastVals[op.index]
}

func (op *objectPath) GetField() reflect.Value {
	return op.Field(op.PathElem().Index)
}

func (op *objectPath) GetIndex() reflect.Value {
	return op.Index(op.PathElem().Index)
}

func (op *objectPath) GetMapValue() reflect.Value {
	return op.MapIndex(op.PathElem().Key)
}

func (op *objectPath) SetMapValueToNew(newType reflect.Type) {
	op.SetMapValue(buildNewValue(newType))
}

func (op *objectPath) SetMapValue(newValue reflect.Value) {
	op.SetMapIndex(op.PathElem().Key, newValue)
}

func (op *objectPath) IsPointer() bool {
	return op.PathElem().Pointer
}

func (op *objectPath) NeedsAppend() bool {
	if op.Len() < op.PathElem().Index {
		panic(NewPatchError("index (%v) larger than slice size(%v)", op.PathElem().Index, op.Len()))
	}
	return op.Len() == op.PathElem().Index
}

func (op *objectPath) AppendNew(newType reflect.Type) {
	op.Append(buildNewValue(newType))
}

func (op *objectPath) Append(newVal reflect.Value) {
	op.Set(reflect.Append(op.Value, newVal))
}

func (op *objectPath) InBounds() bool {
	return op.Len() > op.PathElem().Index
}

func (op *objectPath) CreateIfMissing() {
	if op.IsNil() {
		op.SetToNew(op.Type())
	}
}

func (op *objectPath) SetToNew(newType reflect.Type) {
	op.Set(buildNewValue(newType))
}

func (op *objectPath) Set(newVal reflect.Value) {
	fmt.Println("\n### In set() ###")
	fmt.Printf("CURRENT: %T, settable: %v\n", op.Interface(), op.CanSet())
	fmt.Printf("newVal: %+v\n", newVal)

	settable := op.Value
	prevVal := reflect.ValueOf(nil)
	for i := op.index; !settable.CanSet(); i-- {
		settable = op.lastVals[i]
		prevVal = CopyReflectValue(op.lastVals[i])
		switch prevVal.Kind() {
		case reflect.Struct:
			prevVal.Field(op.Path[i].Index).Set(newVal)
		case reflect.Map:
			prevVal.SetMapIndex(op.Path[i].Key, newVal)
		case reflect.Array:
			fallthrough
		case reflect.Slice:
			prevVal.Index(op.Path[i].Index).Set(newVal)
		case reflect.Ptr:
			prevVal.Elem().Set(newVal)
		default:
			panic(NewPatchError("unhandled set-backtrack kind '%v'\n", prevVal.Kind()))
		}
		newVal = prevVal

		fmt.Printf("CURRENT: %T, settable: %v\n", settable.Interface(), settable.CanSet())
		fmt.Printf("newVal: %+v\n", newVal)
	}

	settable.Set(newVal)
	fmt.Println("### Leaving set() ###")
}

// Delete is only supported for Map, Slice, and Ptr.
func (op *objectPath) Delete() {
	fmt.Println("\n### In delete() ###")
	lastVal := op.LastVal()
	switch lastVal.Kind() {
	case reflect.Map:
		if lastVal.MapIndex(op.Path[op.index].Key).IsValid() {
			op.Set(reflect.Value{})
		}
	case reflect.Slice:
		lastVal.Set(lastVal.Slice(0, lastVal.Len()-1))
	case reflect.Ptr:
		lastVal.Set(reflect.Zero(lastVal.Type()))
	default:
		panic(NewPatchError("unhandled delete kind '%v'", lastVal.Kind()))
	}
	fmt.Println("### Leaving delete() ###")
}

func buildNewValue(newType reflect.Type) (newValue reflect.Value) {
	fmt.Printf("Building new %v\n", newType)
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
	fmt.Printf("Built %+v\n", newValue.Interface())
	return
}
