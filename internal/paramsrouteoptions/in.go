package paramsrouteoptions

type In uint8

const (
	InNone In = iota
	InQuery
	InPath
	InHeader
	InCookie
)
