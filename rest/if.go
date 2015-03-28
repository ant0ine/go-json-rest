package rest

import (
	"log"
)

// IfMiddleware evaluates at runtime a condition based on the current request, and decides to
// execute one of the other Middleware based on this boolean.
type IfMiddleware struct {

	// Runtime condition that decides of the execution of IfTrue of IfFalse.
	Condition func(r *Request) bool

	// Middleware to run when the condition is true. Note that the middleware is initialized
	// weather if will be used or not. (Optional, pass-through if not set)
	IfTrue Middleware

	// Middleware to run when the condition is false. Note that the middleware is initialized
	// weather if will be used or not. (Optional, pass-through if not set)
	IfFalse Middleware
}

// MiddlewareFunc makes TimerMiddleware implement the Middleware interface.
func (mw *IfMiddleware) MiddlewareFunc(h HandlerFunc) HandlerFunc {

	if mw.Condition == nil {
		log.Fatal("IfMiddleware Condition is required")
	}

	var ifTrueHandler HandlerFunc
	if mw.IfTrue != nil {
		ifTrueHandler = mw.IfTrue.MiddlewareFunc(h)
	} else {
		ifTrueHandler = h
	}

	var ifFalseHandler HandlerFunc
	if mw.IfFalse != nil {
		ifFalseHandler = mw.IfFalse.MiddlewareFunc(h)
	} else {
		ifFalseHandler = h
	}

	return func(w ResponseWriter, r *Request) {

		if mw.Condition(r) {
			ifTrueHandler(w, r)
		} else {
			ifFalseHandler(w, r)
		}

	}
}
