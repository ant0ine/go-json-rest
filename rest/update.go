package rest

type Update struct {
	PathExp string
	Func HandlerFunc
}