# xfuego
go-fuego with typed parameters

Package xfuego is an adapter layer around [github.com/go-fuego/fuego](github.com/go-fuego/fuego) route registration functions (Get, Post, etc.)
that adds request parameter typing to request controllers. It supports all the parameter types and options as
provided by fuego: query/path/header/cookie params, bools/ints/strings, optionality, nullability, default values.

Similar to how body type is defined in fuego controller functions, parameters are also defined in controller function
signatures. Parameters are defined in a struct; each struct field is a parameter; parameter options are determined by
each struct field's type and tag:
- example function signature: `func MyController(req xfuego.Request[Params, Body]) (RespBody, error)`
- Base types: bool, int, string 
  - implicitly required unless a default value is provided
- Optional types: *bool, *int, *string
- Nullable types: xfuego.Nullable[bool], xfuego.Nullable[int], xfuego.Nullable[string]
- Optional and nullable types: *xfuego.Nullable[bool], *xfuego.Nullable[int], *xfuego.Nullable[string]
- Parameter tags: `{query,path,header,cookie}:"<name>,<description>,<additional options>"`
  - {query,path,header,cookie} is the parameter `in` value.
  - \<name> is the name of the parameter, if omitted, the struct field name is used.
  - \<additional options> is a comma-separated list of options: `default=<default value>`, `example=<example name>=<example value>`

Package xfuego also introduces the following types:
- `xfuego.Request[Params, Body]` is a wrapper around `fuego.ContextWithBody[Body]` and adds a Params type.
  - request controllers registering through xfuego must use this instead of `fuego.ContextWithBody[Body]`.
- `xfuego.Nullable[T]` is a `*T` that indicates that a parameter is nullable. Null values are represented as `nil`.
- `xfuego.None` is a type that indicates that a request's params and/or body are not used.
  - e.g. `func MyController(req xfuego.Request[xfuego.None, Body]) (RespBody, error)`

Example usage:

```go
package main

import (
	"github.com/crunk1/xfuego"
	"github.com/go-fuego/fuego"
)

func main() {
	s := fuego.NewServer()
	xfuego.Get(s, "/foo/:pathParam/bar", MyController)
	s.Run()
}

type Params struct {
    QueryParam   string `query:"q,My description"` // renamed to "q" with description "My description"
    PathParam    int    `path:"pathParam"`         // renamed to "pathParam", no description
    HeaderParam  bool   `header:",,default=true"`  // default=true (and therefore optional)
    CookieParam  *int   `cookie:""`                // optional (because of pointer) with no default

    WithExamples int `query:",,example=Example1=123,example=Example2=456"` // adds OpenAPI examples: "Example1" and "Example2" as 123 and 456
}

type ReqBody ...

type RespBody ...

func MyController(req xfuego.Request[Params, ReqBody]) (RespBody, error) {
    ...
}
```