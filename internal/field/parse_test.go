package field

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	intT := reflect.TypeOf(0)
	pIntT := reflect.TypeOf((*int)(nil))
	pBoolT := reflect.TypeOf((*bool)(nil))
	stringT := reflect.TypeOf("")

	// Parse consolidates results from parseTag and parseType, so we only need to
	// test the Parse-specific logic here: string conversion, implicit-optionality (defaultValue presence),
	tests := []struct {
		name             string
		fieldType        reflect.Type
		fieldTag         reflect.StructTag
		wantIn           In
		wantRequired     bool
		wantStrconvFn    func(string) any
		wantDefaultValue any
		wantExamples     map[string]any
	}{
		{"no tag", intT, "", InNone, false, nil, nil, nil},
		{"int", intT, `query:""`, InQuery, true, strconvInt, nil, nil},
		{"bool optional", pBoolT, `query:""`, InQuery, false, strconvBool, nil, nil},
		{"int with default (implicit optional)", intT, `query:",,default=1"`, InQuery, false, strconvInt, 1, nil},
		{"int with default (explicit optional)", pIntT, `query:",,default=1"`, InQuery, false, strconvInt, 1, nil},
		{"string with example", stringT, `query:",,example=foo=bar"`, InQuery, true, strconvString, nil, map[string]any{"foo": "bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			field := reflect.StructField{Type: tt.fieldType, Tag: tt.fieldTag}
			gotIn, _, gotRequired, _, gotStrconvFn, _, _, gotDefaultValue, gotExamples := Parse(field)
			a.Equalf(tt.wantIn, gotIn, "Parse(%v)", field)
			a.Equalf(tt.wantRequired, gotRequired, "Parse(%v)", field)
			f1 := runtime.FuncForPC(reflect.ValueOf(tt.wantStrconvFn).Pointer()).Name()
			f2 := runtime.FuncForPC(reflect.ValueOf(gotStrconvFn).Pointer()).Name()
			a.Equalf(f1, f2, "Parse(%v)", field)
			a.Equalf(tt.wantDefaultValue, gotDefaultValue, "Parse(%v)", field)
			a.Equalf(tt.wantExamples, gotExamples, "Parse(%v)", field)
		})
	}
}
