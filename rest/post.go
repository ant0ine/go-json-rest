package rest

type Post struct {
	PathExp string
	Func HandlerFunc
}