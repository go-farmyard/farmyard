package fm

import (
	"github.com/go-farmyard/farmyard/fmutil"
	"reflect"
)

func RequirePointer[T any](c InstanceContainer, optionalName ...string) (ret *T) {
	typ := reflect.TypeOf(ret).Elem()
	if typ.Kind() != reflect.Struct {
		fmutil.Panic("must use a struct type")
	}
	inst, err := GetInstance[*T](c, optionalName...)
	if err != nil {
		fmutil.Panic("failed to require instance pointer: %v", err)
	}
	return inst
}

func RequireInterface[T any](reg InstanceContainer, optionalName ...string) (ret T) {
	typ := reflect.TypeOf(&ret).Elem()
	if typ.Kind() != reflect.Interface {
		fmutil.Panic("must use an interface type")
	}
	inst, err := GetInstance[T](reg, optionalName...)
	if err != nil {
		fmutil.Panic("failed to require instance interface: %v", err)
	}
	return inst
}

func GetInstance[T any](reg InstanceContainer, optionalName ...string) (ret T, err error) {
	name := fmutil.DefZero(optionalName)
	r := reg.(*instanceContainer)
	inst, err := r.getInstance(reflect.TypeOf(&ret).Elem(), name, true)
	if err != nil {
		return ret, err
	}
	return inst.(T), nil
}
