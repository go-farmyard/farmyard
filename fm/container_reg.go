package fm

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"reflect"
)

func (c *instanceContainer) RegisterInstance(inst any, optionalName ...string) {
	name := fmutil.DefZero(optionalName)
	instVal := reflect.ValueOf(inst)
	if instVal.Kind() == reflect.Struct {
		instVal = reflect.ValueOf(toStructPtr(inst))
	} else if instVal.Kind() != reflect.Interface && !(instVal.Kind() == reflect.Ptr && instVal.Elem().Kind() == reflect.Struct) {
		fmutil.Panic("instance must be an interface, struct or a struct pointer, but got: %T", inst)
	}

	instWrapper := &instanceWrapper{inst: instVal.Interface()}
	instType := instVal.Type()
	instTypeStr := instTypeString(instType)

	c.mu.Lock()
	defer c.mu.Unlock()

	if name != "" {
		c.instancesByName[name] = append(c.instancesByName[name], instWrapper)
	}
	c.instancesByType[instTypeStr] = append(c.instancesByType[instTypeStr], instWrapper)
	for i := 0; i < instType.NumMethod(); i++ {
		methodName := instType.Method(i).Name
		c.instancesByMethod[methodName] = append(c.instancesByMethod[methodName], instWrapper)
	}

	instTypeMeta := &instanceTypeMeta{managedFields: map[string]int{}}
	managedFieldStarted := false
	instStructType := instVal.Elem().Type()
	typeAutoInstantiate := reflect.TypeOf(AutoInitialize{})
	for i := 0; i < instStructType.NumField(); i++ {
		field := instStructType.Field(i)
		fieldName := field.Name
		if !managedFieldStarted {
			managedFieldStarted = field.Type == typeAutoInstantiate
			continue
		}
		instTypeMeta.managedFields[fieldName] = i
	}
	instTypeMeta.isManaged = true
	c.instanceTypeMetaMap[instTypeStr] = instTypeMeta
	if managedFieldStarted {
		c.instancesPendingInit = append(c.instancesPendingInit, instVal.Interface())
	}
}
