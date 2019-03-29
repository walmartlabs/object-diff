// Copyright (c) Walmart Inc.
//
// This source code is licensed under the Apache 2.0 license found in the
// LICENSE file in the root directory of this source tree.
package autorec

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

type ApiObject interface {
	metav1.Object
	runtime.Object
}

type Replacer interface {
	ApiObject
	// This retrieves the backing object for this Replacer.
	GetApiObject() ApiObject
	// This sets the backing object for this Replacer.
	SetApiObject(ApiObject)
	// This returns a Replacer implementation specific configuration.
	GetReplacerConfig() interface{}
	// This sets the Replacer implementation specific configuration.
	SetReplacerConfig(interface{}) error
	// This extracts the portion of the ApiObject which should be compared for changes.
	Extract() (interface{}, error)
	// This sets the portion of the ApiObject which should be compared for changes.
	Implant(interface{}) error
}

func NewSimpleReplacer(object ApiObject, field string) Replacer {
	return &simpleReplacer{object, field}
}

func NewDataReplacer(object ApiObject) Replacer {
	return NewSimpleReplacer(object, "Data")
}

func NewSpecReplacer(object ApiObject) Replacer {
	return NewSimpleReplacer(object, "Spec")
}

var _ Replacer = &simpleReplacer{}

type simpleReplacer struct {
	ApiObject
	field string
}

func (sr *simpleReplacer) GetApiObject() ApiObject {
	return sr.ApiObject
}

func (sr *simpleReplacer) SetApiObject(apiObject ApiObject) {
	sr.ApiObject = apiObject
}

func (sr *simpleReplacer) GetReplacerConfig() interface{} {
	return sr.field
}

func (sr *simpleReplacer) SetReplacerConfig(newField interface{}) error {
	field, ok := newField.(string)
	if !ok {
		return fmt.Errorf("trying to set config using %T when string expected", newField)
	}
	sr.field = field
	return nil
}

func (sr *simpleReplacer) Extract() (out interface{}, err error) {
	err = runtime.FieldPtr(reflect.ValueOf(sr.ApiObject).Elem(), sr.field, &out)
	return
}

func (sr *simpleReplacer) Implant(in interface{}) error {
	inTmp := reflect.ValueOf(in).Elem()
	return runtime.SetField(inTmp.Interface(), reflect.ValueOf(sr.ApiObject).Elem(), sr.field)
}
