// Package xfuego is an adapter layer around github.com/go-fuego/fuego route registration functions (Get, Post, etc.)
// that adds request parameter typing to request controllers. It supports all the parameter types and options as
// provided by fuego: query/path/header/cookie params, bools/ints/strings, optionality, nullability, default values.
//
// Similar to how body type is defined in fuego controller functions, parameters are also defined in controller function
// signatures. Parameters are defined in a struct; each struct field is a parameter; parameter options are determined by
// each struct field's type and tag:
//   - example function signature: `func MyController(req xfuego.Request[Params, Body]) (RespBody, error)`
//   - Base types: bool, int, string (required unless a default value is provided)
//   - Optional types: *bool, *int, *string
//   - Nullable types: xfuego.Nullable[bool], xfuego.Nullable[int], xfuego.Nullable[string]
//   - Optional and nullable types: *Nullable[bool], *Nullable[int], *Nullable[string]
//   - Parameter tags: `{query,path,header,cookie}:"<name>,<description>,<additional options>"`
//   - {query,path,header,cookie} is the parameter `in` value.
//   - <name> is the name of the parameter, if omitted, the struct field name is used.
//   - <additional options> is a comma-separated list of options: `default=<default value>`, `example=<example name>=<example value>`
//
// Package xfuego also introduces the following types:
//   - `xfuego.Request[Params, Body]` is a wrapper around `fuego.ContextWithBody[Body]` and adds a Params type.
//   - request controllers registering through xfuego must use this instead of `fuego.ContextWithBody[Body]`.
//   - `xfuego.Nullable[T]` is a `*T` that indicates that a parameter is nullable. Null values are represented as `nil`.
//   - `xfuego.None` is a type that indicates that a request's params and/or body are not used.
//   - e.g. `func MyController(req xfuego.Request[xfuego.None, Body]) (RespBody, error)`
//
// Example usage:
//
//	type Params struct {
//		QueryParam   string `query:"q,My description"` // renamed to "q" with description "My description"
//		PathParam    int    `path:"pathParam"`         // renamed to "pathParam", no description
//		HeaderParam  bool   `header:",,default=true"`  // default=true (and therefore optional)
//		CookieParam  *int   `cookie:""`                // optional (because of pointer) with no default
//
//		WithExamples int		`query:",,example=Example1=123,example=Example2=456"` // adds OpenAPI examples: "Example1" and "Example2" as 123 and 456
//	}
//
//	type ReqBody ...
//
//	type RespBody ...
//
//	func MyController(req xfuego.Request[Params, ReqBody]) (RespBody, error) {
//		...
//	}
package xfuego

import (
	"github.com/go-fuego/fuego"

	"github.com/crunk1/xfuego/internal/paramspopulator"
	"github.com/crunk1/xfuego/internal/paramsrouteoptions"
	"github.com/crunk1/xfuego/internal/types"
)

type Request[ParamsT any, BodyT any] interface {
	fuego.ContextWithBody[BodyT]

	Params() ParamsT
}

type request[ParamsT any, BodyT any] struct {
	fuego.ContextWithBody[BodyT]
	params ParamsT
}

func (r *request[ParamsT, BodyT]) Params() ParamsT {
	return r.params
}

type RequestController[ReqParamsT any, ReqBodyT any, RespBodyT any] func(Request[ReqParamsT, ReqBodyT]) (RespBodyT, error)

type Nullable[T any] = types.Nullable[T]

// None is used to indicate that a request's params and/or body are not used.
type None = types.None

func All[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.All(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Get[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Get(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Post[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Post(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Delete[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Delete(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Put[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Put(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Patch[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Patch(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func Options[ReqParamsT any, ReqBodyT any, RespBodyT any](s *fuego.Server, path string, controller RequestController[ReqParamsT, ReqBodyT, RespBodyT], opts ...func(*fuego.BaseRoute)) *fuego.Route[RespBodyT, ReqBodyT] {
	paramsRouteOptions := paramsrouteoptions.Generate[ReqParamsT]()
	return fuego.Options(s, path, wrapController(controller), append(opts, paramsRouteOptions...)...)
}

func wrapController[ReqParamsT any, ReqBodyT any, RespBodyT any](controller RequestController[ReqParamsT, ReqBodyT, RespBodyT]) func(c fuego.ContextWithBody[ReqBodyT]) (RespBodyT, error) {
	populateParams := paramspopulator.Generate[ReqParamsT]()
	return func(c fuego.ContextWithBody[ReqBodyT]) (RespBodyT, error) {
		req := &request[ReqParamsT, ReqBodyT]{ContextWithBody: c}
		populateParams(c, &req.params)
		return controller(req)
	}
}
