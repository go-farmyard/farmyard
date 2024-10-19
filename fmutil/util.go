package fmutil

import "reflect"

func Ptr[T any](v T) *T {
	return &v
}

func DefZero[T any](defArgs []T) (ret T) {
	return Def(defArgs, ret)
}

func Def[T any](defArgs []T, def T) (ret T) {
	MustTrue(len(defArgs) <= 1, "invalid to have more than one default value")
	if len(defArgs) == 1 {
		return defArgs[0]
	}
	return def
}

func Iif[T any](v bool, t, f T) T {
	if v {
		return t
	}
	return f
}

func IfZero[T any](v T, def T) (ret T) {
	if reflect.DeepEqual(v, ret) {
		return def
	}
	return v
}

func EnsureInterface[T any](_ T) {
}
