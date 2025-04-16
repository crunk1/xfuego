// Package paramspopulator.Generate generates a function that populates a struct with parameters from the request
// context.
//
// Params population logic assumes that validation has already been done before being called.
package paramspopulator

import (
	"net/http"
	"reflect"
	"time"
	"unsafe"

	"github.com/crunk1/xfuego/internal/field"
	"github.com/crunk1/xfuego/internal/types"
)

func Generate[ReqParamsT any]() func(fuegoContextGetters, *ReqParamsT) {
	// No params -> no-op
	if types.IsNoneType[ReqParamsT]() {
		return func(fuegoContextGetters, *ReqParamsT) {}
	}

	t := reflect.TypeOf((*ReqParamsT)(nil)).Elem()

	var populators []func(c fuegoContextGetters, params *ReqParamsT)
	for i := 0; i < t.NumField(); i++ {
		populator := fieldPopulator[ReqParamsT](t.Field(i))
		if populator == nil {
			continue
		}
		populators = append(populators, populator)
	}

	return func(c fuegoContextGetters, params *ReqParamsT) {
		for _, populator := range populators {
			populator(c, params)
		}
	}
}

// fieldPopulator returns a function that populates a field in a Params struct (or nil if the field is not a parameter).
func fieldPopulator[ReqParamsT any](f reflect.StructField) func(c fuegoContextGetters, params *ReqParamsT) {
	in, goKind, _, _, strconvFn, name, _, defaultValue, _ := field.Parse(f)
	if in == field.InNone {
		return nil
	}
	getFieldValueFn := getFns[in]
	setFieldValueFn := setFns[goKind]
	indirectionLevel := getFieldIndirectionLevel(f)

	fieldOffset := f.Offset

	return func(c fuegoContextGetters, params *ReqParamsT) {
		fieldPtr := getFieldPtr(params, fieldOffset)
		valueStr, ok := getFieldValueFn(c, name)
		// If !ok, the field must be optional.
		if !ok {
			if defaultValue != nil {
				setFieldValueFn(fieldPtr, indirectionLevel, defaultValue)
			}
			return
		}
		// "null" handling
		if valueStr == "null" {
			setFieldValueNull(fieldPtr, indirectionLevel)
			return
		}
		// Convert the value to the correct type and set it.
		setFieldValueFn(fieldPtr, indirectionLevel, strconvFn(valueStr))
	}
}

// fuegoContextGetters is a subset of the fuego.ContextWithBody[T] interface that is used to get values from the request.
type fuegoContextGetters interface {
	PathParam(name string) string
	HasQueryParam(name string) bool
	QueryParam(name string) string
	HasHeader(name string) bool
	Header(name string) string
	HasCookie(name string) bool
	Cookie(name string) (*http.Cookie, error)
}

var getFns = map[field.In]func(fuegoContextGetters, string) (string, bool){
	field.InQuery:  getQueryValue,
	field.InPath:   getPathValue,
	field.InHeader: getHeaderValue,
	field.InCookie: getCookieValue,
}

func getPathValue(c fuegoContextGetters, name string) (string, bool) {
	return c.PathParam(name), true
}

func getQueryValue(c fuegoContextGetters, name string) (string, bool) {
	return c.QueryParam(name), c.HasQueryParam(name)
}

func getHeaderValue(c fuegoContextGetters, name string) (string, bool) {
	return c.Header(name), c.HasHeader(name)
}

func getCookieValue(c fuegoContextGetters, name string) (string, bool) {
	if !c.HasCookie(name) {
		return "", false
	}
	cookie, err := c.Cookie(name)
	if err != nil {
		panic(err) // shouldn't happen, Cookie() only errs when cookie DNE
	}
	if cookie.Valid() != nil || (!cookie.Expires.IsZero() && cookie.Expires.Before(time.Now())) {
		return "", false
	}
	return cookie.Value, true
}

var setFns = map[reflect.Kind]func(fieldPtr unsafe.Pointer, indirectionLevel int, value any){
	reflect.Bool:   setFn[bool],
	reflect.Int:    setFn[int],
	reflect.String: setFn[string],
}

func setFn[T any](fieldPtr unsafe.Pointer, indirectionLevel int, value any) {
	v := value.(T)
	if indirectionLevel == 0 {
		*(*T)(fieldPtr) = v
	} else if indirectionLevel == 1 {
		*(**T)(fieldPtr) = &v
	} else if indirectionLevel == 2 {
		pV := &v
		*(***T)(fieldPtr) = &pV
	}
}

// setFieldValueNull is called on `"null"` string values.
// This means that the field is either a Nullable[T] or a *Nullable[T] - 1 or 2 levels of indirection.
func setFieldValueNull(fieldPtr unsafe.Pointer, indirectionLevel int) {
	if indirectionLevel == 1 {
		*(*uintptr)(fieldPtr) = untypedNil
	} else if indirectionLevel == 2 {
		*(**uintptr)(fieldPtr) = &untypedNil
	}
}

var untypedNil = uintptr(0)

// getFieldIndirectionLevel returns the number of levels of indirection for a field.
// A field can have 0-2 levels of indirection:
// - 0: int
// - 1: *int, Nullable[int]
// - 2: *Nullable[int]
func getFieldIndirectionLevel(f reflect.StructField) int {
	// Fields can be 0-2 levels of indirection (e.g. int=0, *int=1, Nullable[int]=1, or *Nullable[int]=2)
	indirectionLevel := 0
	for ft := f.Type; ft.Kind() == reflect.Ptr; ft = ft.Elem() {
		indirectionLevel++
	}
	if indirectionLevel > 2 {
		panic("insanity: field has too many levels of indirection: " + f.Type.String())
	}
	return indirectionLevel
}

func getFieldPtr[T any](params *T, fieldOffset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(params)) + fieldOffset)
}
