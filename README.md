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

Example usage (see `example/main.go`):

```go
/*
Example curls:
curl -X POST "http://localhost:9999/foo/yo/bar?QueryParam=123&nullableParam=null" -d "hello=world"
curl -X POST "http://localhost:9999/foo/yo/bar?QueryParam=123&paramWithDefaultValue=456&optionalParam=false&nullableParam=789&optionalNullableParam=null" -d "hello=world"
curl -X POST "http://localhost:9999/foo/yo/bar?QueryParam=123&optionalParam=true&nullableParam=789&optionalNullableParam=999" -d "hello=world"
curl -X POST -H "OptionalHeaderParam: 123" "http://localhost:9999/foo/yo/bar?QueryParam=123&nullableParam=null" -d "hello=world"
curl -X POST --cookie "OptionalCookieParam=123" "http://localhost:9999/foo/yo/bar?QueryParam=123&nullableParam=null" -d "hello=world"
*/
package main

import (
  "fmt"

  "github.com/go-fuego/fuego"

  "github.com/crunk1/xfuego"
)

func main() {
  s := fuego.NewServer()
  xfuego.Post(s, "/foo/{id}/bar", post)
  s.Run()
}

type Params struct {
  PathParamID           string                `path:"id,My Description"`
  QueryParam            int                   `query:",,example=Ex1=123,example=Ex2=456"`  // name is empty, so defaults to "QueryParam"
  ParamWithDefaultValue int                   `query:"paramWithDefaultValue,,default=123"` // empty description
  OptionalParam         *bool                 `query:"optionalParam"`
  NullableParam         xfuego.Nullable[int]  `query:"nullableParam"`
  OptionalNullableParam *xfuego.Nullable[int] `query:"optionalNullableParam"`

  OptionalHeaderParam *int `header:""`
  OptionalCookieParam *int `cookie:""`
}

type ReqBody struct {
  Hello string
}

type RespBody = string

func post(req xfuego.Request[Params, ReqBody]) (RespBody, error) {
  fmt.Printf("post body: %+v\n", req.MustBody())
  params := req.Params()
  fmt.Println("id path param:", params.PathParamID)
  fmt.Println("QueryParam:", params.QueryParam)
  fmt.Println("ParamWithDefaultValue:", params.ParamWithDefaultValue)
  if params.OptionalParam == nil {
    fmt.Println("optionalParam: <unset>")
  } else {
    fmt.Println("optionalParam:", *params.OptionalParam)
  }
  if params.NullableParam == nil {
    fmt.Println("nullableParam: null")
  } else {
    fmt.Println("nullableParam:", *params.NullableParam)
  }
  if params.OptionalNullableParam == nil {
    fmt.Println("optionalNullableParam: <unset>")
  } else if *params.OptionalNullableParam == nil {
    fmt.Println("optionalNullableParam: null")
  } else {
    fmt.Println("optionalNullableParam: ", **params.OptionalNullableParam)
  }
  if params.OptionalHeaderParam == nil {
    fmt.Println("OptionalHeaderParam: <unset>")
  } else {
    fmt.Println("OptionalHeaderParam:", *params.OptionalHeaderParam)
  }
  if params.OptionalCookieParam == nil {
    fmt.Println("OptionalCookieParam: <unset>")
  } else {
    fmt.Println("OptionalCookieParam:", *params.OptionalCookieParam)
  }
  return "done", nil
}
```