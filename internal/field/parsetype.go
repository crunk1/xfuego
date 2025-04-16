package field

import (
	"reflect"

	"github.com/crunk1/xfuego/internal/types"
)

// parseType determines the Go kind of the field and whether it is required or nullable.
// Field optionality is determined by the presence of a pointer or not, e.g. `*string` vs `string`.
// Field nullability is determined by the presence of a Nullable[T] type, which is also a *T under the hood.
// It is possible that a field is both optional and nullable, e.g. `*Nullable[int]`, so we need to check IsNullable twice.
func parseType(field reflect.StructField) (goKind reflect.Kind, required bool, nullable bool) {
	required = true
	nullable = false
	t := field.Type
	if types.IsNullable(t) {
		nullable = true
		t = t.Elem()
		if t.Kind() == reflect.Pointer {
			panic("param field Nullable type must be a bool|int|string: field=" + field.Name)
		}
	}
	if t.Kind() == reflect.Pointer {
		required = false
		t = t.Elem()
	}
	if types.IsNullable(t) {
		nullable = true
		t = t.Elem()
	}
	goKind = t.Kind()
	if goKind != reflect.Bool && goKind != reflect.Int && goKind != reflect.String {
		panic("param field base type must be a bool|int|string: field=" + field.Name)
	}
	return
}
