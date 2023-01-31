package fmorm

import (
	"fmt"
	"github.com/go-farmyard/farmyard/fmutil"
	"github.com/jmoiron/sqlx/reflectx"
	"reflect"
	"sync"
)

var (
	tpTableName = reflect.TypeOf((*TableNameProvider)(nil)).Elem()
	tvCache     sync.Map
)

func getTableFromSlice(rowsSlicePtr any) (string, bool) {
	sliceValue := reflect.Indirect(reflect.ValueOf(rowsSlicePtr))
	if sliceValue.Kind() != reflect.Slice {
		return "", false
	}
	sliceElementType := sliceValue.Type().Elem()
	var pv reflect.Value
	if sliceElementType.Kind() == reflect.Ptr {
		pv = reflect.New(sliceElementType.Elem())
	} else if sliceElementType.Kind() == reflect.Struct {
		pv = reflect.New(sliceElementType)
	} else {
		return "", false
	}
	return getTableName(pv), true
}

func getTableName(v reflect.Value) string {
	if v.Type().Implements(tpTableName) {
		return v.Interface().(TableNameProvider).TableName()
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		if v.Type().Implements(tpTableName) {
			return v.Interface().(TableNameProvider).TableName()
		}
	} else if v.CanAddr() {
		v1 := v.Addr()
		if v1.Type().Implements(tpTableName) {
			return v1.Interface().(TableNameProvider).TableName()
		}
	} else {
		name, ok := tvCache.Load(v.Type())
		if ok {
			if tableName, ok := name.(string); ok && tableName != "" {
				return tableName
			}
		} else {
			v2 := reflect.New(v.Type())
			if v2.Type().Implements(tpTableName) {
				tableName := v2.Interface().(TableNameProvider).TableName()
				tvCache.Store(v.Type(), tableName)
				return tableName
			}
			tvCache.Store(v.Type(), "")
		}
	}

	return v.Type().Name()
}

func convertMapStringInterface(v any) (map[string]any, bool) {
	var m map[string]any
	mtype := reflect.TypeOf(m)
	t := reflect.TypeOf(v)
	if !t.ConvertibleTo(mtype) {
		return nil, false
	}
	return reflect.ValueOf(v).Convert(mtype).Interface().(map[string]any), true
}

type ModelStructField struct {
	Name  string
	Value any
}

type ModelStructFields []*ModelStructField

func modelStructFieldValues(fieldMapper *reflectx.Mapper, model any, full bool) (ModelStructFields, error) {
	var fields ModelStructFields
	err := modelStructFieldTraversal(fieldMapper, model, nil, func(name string, val reflect.Value) {
		if full || !val.IsZero() {
			fields = append(fields, &ModelStructField{Name: name, Value: val.Interface()})
		}
	})
	return fields, err
}

func modelStructFieldTraversal(fieldMapper *reflectx.Mapper, model any, names []string, fn func(string, reflect.Value)) error {
	var v reflect.Value
	for v = reflect.ValueOf(model); v.Kind() == reflect.Ptr; {
		v = v.Elem()
	}
	if v.Type().Kind() == reflect.Map {
		var m map[string]any
		mType := reflect.TypeOf(m)
		if v.Type().ConvertibleTo(mType) {
			m = reflect.ValueOf(v).Convert(mType).Interface().(map[string]any)
			if names == nil {
				for mk, mv := range m {
					fn(mk, reflect.ValueOf(mv))
				}
			} else {
				for _, name := range names {
					if mv, ok := m[name]; ok {
						fn(name, reflect.ValueOf(mv))
					} else {
						return fmt.Errorf("could not find name %s in map %#v", name, model)
					}
				}
			}
			return nil
		}
	}
	if v.Type().Kind() != reflect.Struct {
		return fmt.Errorf("unsupported model %#v", model)
	}

	typeMap := fieldMapper.TypeMap(v.Type())
	if names != nil {
		return fieldMapper.TraversalsByNameFunc(v.Type(), names, func(i int, t []int) error {
			if len(t) == 0 {
				return fmt.Errorf("could not find name %s in model %#v", names[i], model)
			}
			val := reflectx.FieldByIndexesReadOnly(v, t)
			fn(names[i], val)
			return nil
		})
	} else {
		for _, idx := range typeMap.Index {
			val := reflectx.FieldByIndexesReadOnly(v, idx.Index)
			fn(idx.Name, val)
		}
	}
	return nil
}

func interfaceSlice(slice any) []any {
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("expect to get a slice")
	}
	if s.IsNil() {
		return nil
	}
	ret := make([]any, s.Len())
	for i := 0; i < s.Len(); i++ {
		ret[i] = s.Index(i).Interface()
	}
	return ret
}

func resetToZeroValue(v any) {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr {
		fmutil.Panic("the value must be a pointer to be reset to zero, but got: %T", v)
	}
	if !val.IsZero() {
		typ := val.Elem().Type()
		val.Elem().Set(reflect.New(typ).Elem())
	}
}
