package types

import (
	"reflect"
)

// None is used to indicate that a request's params and/or body are not used.
type None any

var nonePtrType = reflect.TypeOf((*None)(nil))

func IsNoneType[T any]() bool {
	return reflect.TypeOf((*T)(nil)) == nonePtrType
}
