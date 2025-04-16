package paramsrouteoptions

import (
	"reflect"

	"github.com/go-fuego/fuego"

	"github.com/crunk1/xfuego/internal/field"
	"github.com/crunk1/xfuego/internal/types"
)

func Generate[ReqParamsT any]() []func(*fuego.BaseRoute) {
	// No params -> no-op
	if types.IsNoneType[ReqParamsT]() {
		return nil
	}

	// Check params is a struct.
	t := reflect.TypeOf((*ReqParamsT)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		panic("ReqParamsT type must be a struct: type=" + t.String())
	}

	var opts []func(*fuego.BaseRoute)
	for i := 0; i < t.NumField(); i++ {
		opt := parsedFieldToRouteOption(field.Parse(t.Field(i)))
		if opt == nil {
			continue
		}
		opts = append(opts, opt)
	}
	return opts
}

func parsedFieldToRouteOption(in field.In, goKind reflect.Kind, required bool, nullable bool, _ func(string) any, name string, desc string, defaultValue any, examples map[string]any) func(*fuego.BaseRoute) {
	if in == field.InNone {
		return nil
	}

	// param opts: required, default, examples, nullable
	var paramOpts []func(param *fuego.OpenAPIParam)
	if required {
		paramOpts = append(paramOpts, fuego.ParamRequired())
	} else if defaultValue != nil {
		paramOpts = append(paramOpts, fuego.ParamDefault(defaultValue))
	}
	if nullable {
		paramOpts = append(paramOpts, fuego.ParamNullable())
	}
	for exampleName, exampleValue := range examples {
		paramOpts = append(paramOpts, fuego.ParamExample(exampleName, exampleValue))
	}

	// Query options. Has special handling for types.
	if in == field.InQuery {
		if goKind == reflect.String {
			return fuego.OptionQuery(name, desc, paramOpts...)
		} else if goKind == reflect.Int {
			return fuego.OptionQueryInt(name, desc, paramOpts...)
		} else if goKind == reflect.Bool {
			return fuego.OptionQueryBool(name, desc, paramOpts...)
		}
	}

	// Path, Header, Cookie options.
	if goKind == reflect.String {
		paramOpts = append(paramOpts, fuego.ParamString())
	} else if goKind == reflect.Int {
		paramOpts = append(paramOpts, fuego.ParamInteger())
	} else if goKind == reflect.Bool {
		paramOpts = append(paramOpts, fuego.ParamBool())
	}
	if in == field.InPath {
		return fuego.OptionPath(name, desc, paramOpts...)
	} else if in == field.InHeader {
		return fuego.OptionHeader(name, desc, paramOpts...)
	} else if in == field.InCookie {
		return fuego.OptionCookie(name, desc, paramOpts...)
	}

	// Shouldn't reach here, but I wanted to be explicit in the if statements above - i.e. no catch-all `else` case
	return nil
}
