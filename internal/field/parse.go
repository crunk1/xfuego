package field

import (
	"reflect"
	"strconv"
)

// Parse parses a field's type and tag information
func Parse(field reflect.StructField) (in In, goKind reflect.Kind, required bool, nullable bool, strconvFn func(string) any, name string, desc string, defaultValue any, examples map[string]any) {
	in, name, desc, defaultValue, examples = parseTag(field)
	if in == InNone {
		return InNone, 0, false, false, nil, "", "", nil, nil
	}
	if !field.IsExported() {
		panic("param field must be exported: field=" + field.Name)
	}
	if field.Anonymous { // TODO: support public anonymous fields - embedded structs
		panic("param anonymous field support is not yet implemented: field=" + field.Name)
	}
	goKind, required, nullable = parseType(field)

	// Name defaulting
	if name == "" {
		name = field.Name
	}

	// Not required if defaultValue is set
	required = defaultValue == nil && required

	// Set the string conversion function based on the field type.
	if goKind == reflect.Bool {
		strconvFn = strconvBool
	} else if goKind == reflect.Int {
		strconvFn = strconvInt
	} else if goKind == reflect.String {
		strconvFn = strconvString
	}

	// Convert defaultValue and example strings.
	if defaultValue != nil {
		defaultValue = strconvFn(defaultValue.(string))
	}
	for exampleName, exampleValue := range examples {
		examples[exampleName] = strconvFn(exampleValue.(string))
	}

	return
}

func strconvBool(value string) any {
	result, err := strconv.ParseBool(value)
	if err != nil {
		panic("param string value is not a bool: " + value)
	}
	return result
}

func strconvInt(value string) any {
	result, err := strconv.Atoi(value)
	if err != nil {
		panic("param string value is not an int: " + value)
	}
	return result
}

func strconvString(value string) any {
	return value
}
