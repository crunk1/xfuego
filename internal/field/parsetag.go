package field

import (
	"reflect"
	"strings"
)

// parseTag parses the struct tag for a parameter and returns the location and tag value.
// It supports the following tags: query, path, header, and cookie.
//
// A valid tag value is of the form:
// "name,description,default=foo,example=exampleName=foo,example=exampleName2=bar"
func parseTag(field reflect.StructField) (in In, name string, desc string, defaultValue any, examples map[string]any) {
	matches := 0
	var tagValue string
	tag := field.Tag
	if queryTag, ok := tag.Lookup("query"); ok {
		in = InQuery
		tagValue = queryTag
		matches++
	}
	if pathTag, ok := tag.Lookup("path"); ok {
		in = InPath
		tagValue = pathTag
		matches++
	}
	if headerTag, ok := tag.Lookup("header"); ok {
		in = InHeader
		tagValue = headerTag
		matches++
	}
	if cookieTag, ok := tag.Lookup("cookie"); ok {
		in = InCookie
		tagValue = cookieTag
		matches++
	}
	if matches > 1 {
		panic("param field cannot have more than one param tag: field=" + field.Name)
	}
	name, desc, defaultValue, examples = parseTagValue(tagValue)
	return
}

// parseTagValue breaks down the param tag value into its components.
//
// A valid tag value is of the form:
//
//	"name,description,default=foo,example=exampleName=foo,example=exampleName2=bar"
func parseTagValue(tagValue string) (name, desc string, defaultValue any, examples map[string]any) {
	parts := strings.Split(tagValue, ",")
	if len(parts) >= 1 {
		name = parts[0]
	}
	if len(parts) >= 2 {
		desc = parts[1]
	}
	if len(parts) < 3 {
		return name, desc, nil, nil
	}
	parts = parts[2:]

	// param opts: default, example, required
	for _, part := range parts {
		optParts := strings.SplitN(part, "=", 2)
		if optParts[0] == "default" {
			if len(optParts) == 1 {
				panic("param opt 'default' must have a value, param opts: " + tagValue)
			}
			defaultValue = optParts[1]
		} else if optParts[0] == "example" {
			if len(optParts) == 1 {
				panic("param opt 'example' must have a value, param opts: " + tagValue)
			}
			exampleParts := strings.SplitN(optParts[1], "=", 2)
			if len(exampleParts) == 1 {
				panic("param opt 'example' must be a 'key=value' string, param opts: " + tagValue)
			}
			if examples == nil {
				examples = make(map[string]any)
			}
			examples[exampleParts[0]] = exampleParts[1]
		} else {
			panic("unknown param opt '" + optParts[0] + "', param opts: " + tagValue)
		}
	}

	return name, desc, defaultValue, examples
}
