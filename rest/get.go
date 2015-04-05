package rest

type Get struct {
	PathExp string
	Func HandlerFunc
}