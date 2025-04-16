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
