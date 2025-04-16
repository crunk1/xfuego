package field

// In is used to indicate where a parameter is located in the request.
type In uint8

const (
	InNone In = iota
	InQuery
	InPath
	InHeader
	InCookie
)
