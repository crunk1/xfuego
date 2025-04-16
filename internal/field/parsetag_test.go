package field

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Parse(t *testing.T) {
	tests := []struct {
		name   string
		argTag reflect.StructTag
		wantIn In
		// Leave these to Test_parseTagValue
		// wantName         string
		// wantDesc         string
		// wantDefaultValue *string
		// wantExamples     map[string]string
		wantPanic bool
	}{
		{"empty tag", "", InNone, false},
		{"other tags", `json:"foo"`, InNone, false},
		{"query tag", `query:"foo"`, InQuery, false},
		{"path tag", `path:"foo"`, InPath, false},
		{"header tag", `header:"foo"`, InHeader, false},
		{"cookie tag", `cookie:"foo"`, InCookie, false},
		{"panic on multiple tags", `query:"foo" path:"bar"`, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			field := reflect.StructField{Tag: tt.argTag}
			if tt.wantPanic {
				a.Panics(func() { parseTag(field) })
				return
			}
			gotIn, _, _, _, _ := parseTag(field)
			if gotIn != tt.wantIn {
				t.Errorf("gotIn %v, wantIn %v", gotIn, tt.wantIn)
			}
		})
	}
}

func Test_parseTagValue(t *testing.T) {
	tests := []struct {
		name             string
		argTagValue      string
		wantName         string
		wantDesc         string
		wantDefaultValue any
		wantExamples     map[string]any
	}{
		{"name only", "name", "name", "", nil, nil},
		{"name and desc", "name,description", "name", "description", nil, nil},
		{"desc but no name", ",description", "", "description", nil, nil},
		{"name and default", "name,,default=foo", "name", "", "foo", nil},
		{"example", ",,example=exampleName=foo", "", "", nil, map[string]any{"exampleName": "foo"}},
		{"example,default,example", ",,example=exampleName=foo,default=bar,example=exampleName2=foo2", "", "", "bar", map[string]any{"exampleName": "foo", "exampleName2": "foo2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			gotName, gotDesc, gotDefaultValue, gotExamples := parseTagValue(tt.argTagValue)
			a.Equal(tt.wantName, gotName)
			a.Equal(tt.wantDesc, gotDesc)
			a.Equal(tt.wantDefaultValue, gotDefaultValue)
			a.Equal(tt.wantExamples, gotExamples)
		})
	}
}
