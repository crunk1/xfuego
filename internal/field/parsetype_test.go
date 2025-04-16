package field

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/crunk1/xfuego/internal/types"
)

func Test_parseType(t *testing.T) {
	intT := reflect.TypeOf(0)
	pIntT := reflect.TypeOf((*int)(nil))
	ppIntT := reflect.TypeOf((**int)(nil))
	nullableIntT := reflect.TypeOf((*types.Nullable[int])(nil)).Elem()
	pNullableIntT := reflect.TypeOf((**types.Nullable[int])(nil)).Elem()
	nullablePIntT := reflect.TypeOf((*types.Nullable[*int])(nil)).Elem()

	tests := []struct {
		name         string
		fieldType    reflect.Type
		wantGoKind   reflect.Kind
		wantRequired bool
		wantNullable bool
		wantPanic    bool
	}{
		{"basic int", intT, reflect.Int, true, false, false},
		{"optional int - *int", pIntT, reflect.Int, false, false, false},
		{"nullable int - Nullable[int]", nullableIntT, reflect.Int, true, true, false},
		{"optional nullable int - *Nullable[int]", pNullableIntT, reflect.Int, false, true, false},
		{"bad - **int", ppIntT, reflect.Int, false, false, true},
		{"bad - Nullable[*int]", nullablePIntT, reflect.Int, false, true, true},
		{"bad non{bool,int,string} - []int", reflect.SliceOf(intT), reflect.Int, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			field := reflect.StructField{Type: tt.fieldType}
			if tt.wantPanic {
				a.Panics(func() { parseType(field) })
				return
			}
			gotGoKind, gotRequired, gotNullable := parseType(field)
			a.Equalf(tt.wantGoKind, gotGoKind, "parseType(%v)", field)
			a.Equalf(tt.wantRequired, gotRequired, "parseType(%v)", field)
			a.Equalf(tt.wantNullable, gotNullable, "parseType(%v)", field)
		})
	}
}
