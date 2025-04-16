package paramspopulator

import (
	"net/http"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"

	"github.com/crunk1/xfuego/internal/types"
)

func TestGenerate(t *testing.T) {
	type Params struct {
		Field0 int                  `query:""`
		Field1 *int                 `query:""`
		Field2 int                  `query:",,default=1"`
		Field3 types.Nullable[int]  `query:""`
		Field4 *types.Nullable[int] `query:""`
	}

	type testCase struct {
		name        string
		queryParams map[string]string
		want        Params
	}
	tests := []testCase{
		{"case0", map[string]string{"Field0": "1", "Field3": "null"}, Params{Field0: 1, Field2: 1}},
		{"case1", map[string]string{"Field0": "1", "Field1": "1", "Field3": "null"}, Params{Field0: 1, Field1: lo.ToPtr(1), Field2: 1}},
		{"case2", map[string]string{"Field0": "1", "Field2": "2", "Field3": "null"}, Params{Field0: 1, Field2: 2}},
		{"case3", map[string]string{"Field0": "1", "Field3": "1"}, Params{Field0: 1, Field2: 1, Field3: types.Nullable[int](lo.ToPtr(1))}},
		{"case4", map[string]string{"Field0": "1", "Field3": "null", "Field4": "null"}, Params{Field0: 1, Field2: 1, Field4: lo.ToPtr(types.Nullable[int](nil))}},
		{"case5", map[string]string{"Field0": "1", "Field3": "null", "Field4": "1"}, Params{Field0: 1, Field2: 1, Field4: lo.ToPtr(types.Nullable[int](lo.ToPtr(1)))}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			populate := Generate[Params]()
			getters := &mockGetters{query: tt.queryParams}
			params := &Params{}
			populate(getters, params)
			a.Equal(tt.want, *params)
		})
	}
}

func Test_fieldPopulator(t *testing.T) {
	type Params struct {
		Field0 int                  `path:"foo"`
		Field1 *int                 `query:"unset_optional"`
		Field2 *int                 `query:"set_optional"`
		Field3 int                  `query:"unset_default,,default=123"`
		Field4 types.Nullable[int]  `query:"null_nullable"`
		Field5 *types.Nullable[int] `query:"unset_optional_nullable"`
		Field6 *types.Nullable[int] `query:"set_optional_nullable"`
		Field7 *types.Nullable[int] `query:"set_optional_null_nullable"`
	}
	paramsT := reflect.TypeOf((*Params)(nil)).Elem()

	getters := &mockGetters{
		path: map[string]string{"foo": "123"},
		query: map[string]string{
			"set_optional":               "123",
			"null_nullable":              "null",
			"set_optional_nullable":      "123",
			"set_optional_null_nullable": "null",
		},
	}

	type testCase struct {
		name       string
		fieldIndex int
		want       any
	}
	tests := []testCase{
		{"required field", 0, 123},
		{"unset optional field", 1, (*int)(nil)},
		{"set optional field", 2, lo.ToPtr(123)},
		{"default value", 3, 123},
		{"null value", 4, types.Nullable[int](nil)},
		{"unset optional nullable field", 5, (*types.Nullable[int])(nil)},
		{"set optional nullable field", 6, lo.ToPtr(types.Nullable[int](lo.ToPtr(123)))},
		{"set optional null nullable field", 7, lo.ToPtr(types.Nullable[int](nil))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			params := Params{}
			paramsV := reflect.ValueOf(&params).Elem()
			field := paramsT.Field(tt.fieldIndex)
			gotFn := fieldPopulator[Params](field)
			a.NotNil(gotFn)
			gotFn(getters, &params)
			a.Equal(tt.want, paramsV.Field(tt.fieldIndex).Interface())
		})
	}
}

func Test_getCookieValue(t *testing.T) {
	getters := &mockGetters{
		cookies: map[string]*http.Cookie{
			"foo":        {Name: "foo", Value: "bar"},
			"expired":    {Name: "expired", Value: "expired", Expires: time.Now().Add(-time.Hour)},
			"notexpired": {Name: "notexpired", Value: "notexpired", Expires: time.Now().Add(time.Hour)},
			"invalid":    {Name: "invalid", Value: `¯\_(ツ)_/¯`}, // invalid value character: ツ
		},
	}

	tests := []struct {
		name     string
		argsName string
		want     string
		wantOk   bool
	}{
		{"exists", "foo", "bar", true},
		{"not exists", "baz", "", false},
		{"expired cookie", "expired", "", false},
		{"not expired cookie", "notexpired", "notexpired", true},
		{"invalid cookie", "invalid", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			got, gotOk := getCookieValue(getters, tt.argsName)
			a.Equal(tt.want, got)
			a.Equal(tt.wantOk, gotOk)
		})
	}
}

func Test_getFieldIndirectionLevel(t *testing.T) {
	type Params struct {
		Field0 int
		Field1 *int
		Field2 types.Nullable[int]
		Field3 *types.Nullable[int]
		Field4 **types.Nullable[int]
		Field5 ***int
	}
	paramsT := reflect.TypeOf((*Params)(nil)).Elem()

	tests := []struct {
		name      string
		f         reflect.StructField
		want      int
		wantPanic bool
	}{
		{"0 indirection", paramsT.Field(0), 0, false},
		{"1 indirection - *int", paramsT.Field(1), 1, false},
		{"1 indirection - Nullable[int]", paramsT.Field(2), 1, false},
		{"2 indirection - *Nullable[int]", paramsT.Field(3), 2, false},
		{"panic - **Nullable[int]", paramsT.Field(4), 0, true},
		{"panic - ***int", paramsT.Field(5), 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			if tt.wantPanic {
				a.Panics(func() {
					getFieldIndirectionLevel(tt.f)
				}, "getFieldIndirectionLevel() should panic")
				return
			}
			if got := getFieldIndirectionLevel(tt.f); got != tt.want {
				t.Errorf("getFieldIndirectionLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getHeaderValue(t *testing.T) {
	getters := &mockGetters{headers: map[string]string{"foo": "bar"}}

	tests := []struct {
		name     string
		argsName string
		want     string
		wantOk   bool
	}{
		{"exists", "foo", "bar", true},
		{"not exists", "baz", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			got, gotOk := getHeaderValue(getters, tt.argsName)
			a.Equal(tt.want, got)
			a.Equal(tt.wantOk, gotOk)
		})
	}
}

func Test_getPathValue(t *testing.T) {
	getters := &mockGetters{path: map[string]string{"foo": "bar"}}

	tests := []struct {
		name     string
		argsName string
		want     string
	}{
		{"exists", "foo", "bar"},
		// {"not exists", "baz", ""}, // Path params are always required, and thus should always be present
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			got, gotOk := getPathValue(getters, tt.argsName)
			a.Equal(tt.want, got)
			a.True(gotOk)
		})
	}
}

func Test_getQueryValue(t *testing.T) {
	getters := &mockGetters{query: map[string]string{"foo": "bar"}}

	tests := []struct {
		name     string
		argsName string
		want     string
		wantOk   bool
	}{
		{"exists", "foo", "bar", true},
		{"not exists", "baz", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			got, gotOk := getQueryValue(getters, tt.argsName)
			a.Equal(tt.want, got)
			a.Equal(tt.wantOk, gotOk)
		})
	}
}

func Test_setFieldValueNull(t *testing.T) {
	type Params struct {
		Field0 types.Nullable[int]
		Field1 *types.Nullable[int]
	}
	paramsT := reflect.TypeOf((*Params)(nil)).Elem()

	tests := []struct {
		name                 string
		argsFieldIndex       int
		argsIndirectionLevel int
	}{
		{"Nullable[int]", 0, 1},
		{"*Nullable[int]", 1, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			// Init params with non-null values to test nullification
			params := Params{Field0: types.Nullable[int](lo.ToPtr(123)), Field1: lo.ToPtr(types.Nullable[int](lo.ToPtr(123)))}
			paramsV := reflect.ValueOf(&params).Elem()
			fieldPtr := getFieldPtr(&params, paramsT.Field(tt.argsFieldIndex).Offset)
			setFieldValueNull(fieldPtr, tt.argsIndirectionLevel)
			if tt.argsIndirectionLevel == 1 {
				a.Nil(paramsV.Field(tt.argsFieldIndex).Interface().(types.Nullable[int]))
			} else if tt.argsIndirectionLevel == 2 {
				a.Nil(*paramsV.Field(tt.argsFieldIndex).Interface().(*types.Nullable[int]))
			}
		})
	}
}

func Test_setFn(t *testing.T) {
	type Params struct {
		Field0 int
		Field1 *int
		Field2 types.Nullable[int]
		Field3 *types.Nullable[int]
	}
	paramsT := reflect.TypeOf((*Params)(nil)).Elem()

	tests := []struct {
		name             string
		fieldIndex       int
		indirectionLevel int
		value            any
		want             int
	}{
		{"int", 0, 0, 123, 123},
		{"*int", 1, 1, 123, 123},
		{"Nullable[int]", 2, 1, 123, 123},
		{"*Nullable[int]", 3, 2, 123, 123},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			params := &Params{}
			paramsElemV := reflect.ValueOf(params).Elem()
			field := paramsT.Field(tt.fieldIndex)
			fieldPtr := getFieldPtr(params, field.Offset)
			setFn[int](fieldPtr, tt.indirectionLevel, tt.value)
			if tt.indirectionLevel == 0 {
				a.Equal(tt.want, paramsElemV.Field(tt.fieldIndex).Interface())
			} else if tt.indirectionLevel == 1 {
				got := *(*int)(paramsElemV.Field(tt.fieldIndex).UnsafePointer())
				a.Equal(tt.want, got)
			} else if tt.indirectionLevel == 2 {
				got := **(**int)(paramsElemV.Field(tt.fieldIndex).UnsafePointer())
				a.Equal(tt.want, got)
			}
		})
	}
}

func Test_getFieldPtr(t *testing.T) {
	type Params struct {
		Bool bool
		Int  int
		Str  string
	}
	paramsT := reflect.TypeOf((*Params)(nil)).Elem()
	params := &Params{}

	type testCase struct {
		name       string
		fieldIndex int
		want       unsafe.Pointer
	}
	tests := []testCase{
		{"first field", 0, unsafe.Pointer(params)},
		{"second field", 1, unsafe.Pointer(uintptr(unsafe.Pointer(params)) + unsafe.Offsetof(params.Int))},
		{"third field", 2, unsafe.Pointer(uintptr(unsafe.Pointer(params)) + unsafe.Offsetof(params.Str))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			field := paramsT.Field(tt.fieldIndex)
			fieldOffset := field.Offset
			got := getFieldPtr(params, fieldOffset)
			a.Equal(uintptr(tt.want), uintptr(got), "getFieldPtr(%v, %v)", params, fieldOffset)
		})
	}
}

type mockGetters struct {
	cookies map[string]*http.Cookie
	headers map[string]string
	path    map[string]string
	query   map[string]string
}

func (mg *mockGetters) Cookie(name string) (*http.Cookie, error) {
	return mg.cookies[name], nil
}

func (mg *mockGetters) HasCookie(name string) bool {
	return mg.cookies[name] != nil
}

func (mg *mockGetters) Header(name string) string {
	return mg.headers[name]
}

func (mg *mockGetters) HasHeader(name string) bool {
	return mg.headers[name] != ""
}

func (mg *mockGetters) PathParam(name string) string {
	return mg.path[name]
}

func (mg *mockGetters) QueryParam(name string) string {
	return mg.query[name]
}

func (mg *mockGetters) HasQueryParam(name string) bool {
	return mg.query[name] != ""
}
