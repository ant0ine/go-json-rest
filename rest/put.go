package rest

type Put struct {
	PathExp string
	Func HandlerFunc
}