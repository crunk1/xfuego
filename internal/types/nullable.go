package types

import (
	"reflect"
	"strings"
)

type Nullable[T any] *T

// IsNullable checks if the given type is a Nullable[T] type.
// Unfortunately, as of Go 1.24.2, we have to rely on string comparison to check if a type is a Nullable[T] type.
func IsNullable(t reflect.Type) bool {
	tFullName := t.PkgPath() + "." + t.Name()
	return strings.HasPrefix(tFullName, nullableTypeFullNameSansTypeParams)
}

var nullableType = reflect.TypeOf((*Nullable[any])(nil)).Elem()
var nullableTypeFullName = nullableType.PkgPath() + "." + nullableType.Name()                                // "package/path/types.Nullable[interface {}]"
var nullableTypeFullNameSansTypeParams = nullableTypeFullName[:strings.LastIndex(nullableTypeFullName, "[")] // "package/path/types.Nullable"
