package paramsrouteoptions

import (
	"reflect"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/go-fuego/fuego"
	"github.com/stretchr/testify/assert"

	"github.com/crunk1/xfuego/internal/field"
	"github.com/crunk1/xfuego/internal/types"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name        string
		genFn       func() []func(*fuego.BaseRoute)
		wantOptsLen int
		wantPanic   bool
	}{
		{"None", Generate[types.None], 0, false},
		{"empty struct", Generate[struct{}], 0, false},
		{"struct with query", Generate[struct {
			X int `query:""`
		}], 1, false},
		{"struct with nonparam field", Generate[struct {
			X int `query:""`
			Y int `json:"y"`
		}], 1, false},
		{"panic on non-struct (int)", Generate[int], 0, true},
		{"panic on non-struct (*struct)", Generate[*struct{}], 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			if tt.wantPanic {
				a.Panics(func() { tt.genFn() })
				return
			}
			gotOpts := tt.genFn()
			if len(gotOpts) != tt.wantOptsLen {
				t.Errorf("got %v, want %v", len(gotOpts), tt.wantOptsLen)
			}
		})
	}
}

func Test_parsedFieldToRouteOption(t *testing.T) {
	type args struct {
		in       field.In
		goKind   reflect.Kind
		required bool
		nullable bool
		// strconvFn    func(string) any // not used by parsedFieldToRouteOption
		// name         string           // permanently set as "x"
		// desc         string           // permanently set as "x description"
		defaultValue any
		examples     map[string]any
	}
	const argsName = "x"
	const argsDesc = "x description"
	tests := []struct {
		name      string
		args      args
		wantParam *fuego.OpenAPIParam // nil is used to indicate parsedFieldToRouteOption should return nil
		wantPanic bool
	}{
		{
			"not a param",
			args{in: field.InNone},
			nil,
			false,
		},
		{
			"path - basic int",
			args{in: field.InPath, goKind: reflect.Int},
			&fuego.OpenAPIParam{Type: "path", Required: true, GoType: "integer"}, // path params are always required
			false,
		},
		{
			"path - basic int with default value",
			args{in: field.InPath, goKind: reflect.Int, defaultValue: 10},
			&fuego.OpenAPIParam{Type: "path", Required: true, GoType: "integer", Default: 10}, // path params are always required
			false,
		},
		{
			"path - basic int with example",
			args{in: field.InPath, goKind: reflect.Int, examples: map[string]any{"123": 456}},
			&fuego.OpenAPIParam{Type: "path", Required: true, GoType: "integer", Examples: map[string]any{"123": 456}}, // path params are always required
			false,
		},
		{
			"query - required int",
			args{in: field.InQuery, goKind: reflect.Int, required: true},
			&fuego.OpenAPIParam{Type: "query", Required: true, GoType: "integer"},
			false,
		},
		{
			"query - optional int",
			args{in: field.InQuery, goKind: reflect.Int},
			&fuego.OpenAPIParam{Type: "query", GoType: "integer"},
			false,
		},
		{
			"query - nullable",
			args{in: field.InQuery, goKind: reflect.Int, nullable: true, required: true},
			&fuego.OpenAPIParam{Type: "query", Required: true, GoType: "integer", Nullable: true},
			false,
		},
		{
			"query - optional nullable",
			args{in: field.InQuery, goKind: reflect.Int, nullable: true},
			&fuego.OpenAPIParam{Type: "query", GoType: "integer", Nullable: true},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := assert.New(t)
			args := &tt.args
			if tt.wantPanic {
				a.Panics(func() {
					parsedFieldToRouteOption(args.in, args.goKind, args.required, args.nullable, nil, argsName, argsDesc, args.defaultValue, args.examples)
				})
				return
			}
			routeOpt := parsedFieldToRouteOption(args.in, args.goKind, args.required, args.nullable, nil, argsName, argsDesc, args.defaultValue, args.examples)
			if tt.wantParam == nil {
				a.Nil(routeOpt)
				return
			}
			a.NotNil(routeOpt)
			tt.wantParam.Name = argsName
			tt.wantParam.Description = argsDesc
			route := &fuego.BaseRoute{Operation: &openapi3.Operation{}}
			if routeOpt != nil {
				routeOpt(route)
			}
			a.Equal(*tt.wantParam, route.Params[argsName])
		})
	}
}
