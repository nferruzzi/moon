// Middleware chains for Go web applications with Context from x/net/context and http.Handler support
package moon

import (
	"net/http"

	"golang.org/x/net/context"
)

// Middleware interface, returns a standard http.HandlerFunc
// Is used to declare new middlewares that can be passed to moon.New
// Inside you can call Next with the same or a new context to go
// ahead with the next middleware
type Middleware func(context.Context, Next) http.Handler

// Handler interface, return a standar http.Handler
// Is called at the end of the middleware chain to handle the request/response
type Handler func(context.Context) http.Handler

// Next interface, is used for advance to the next Middleware
type Next func(context.Context)

// HandlerFunc interface is a shortcut to get a function called
// at the end of the middleware chain
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request)

// Moon stores the middlewares. Use moon.New to create a new one.
type Moon struct {
	mws []Middleware
}

var (
	// Callback to setup the root context for each request, default is: context.TODO()
	// Appengine should use appengine.NewContext(request)
	Context func(r *http.Request) context.Context
)

// New is used to create the chain of middlewares; to append the final handler you
// call ThenFunc or Then
func New(middlewares ...Middleware) Moon {
	return Moon{mws: middlewares}
}

// Internal function; it's purpose is to iterate the chain and call the relative handlers till the end
// where it's called the end Handler
func (moon Moon) runMiddleware(i int, ctx context.Context, w http.ResponseWriter, r *http.Request, end Handler) {
	if i == len(moon.mws) {
		// end recursion
		if end != nil {
			end(ctx).ServeHTTP(w, r)
		}
	} else {
		// configure call to next middleware
		next := func(ctx context.Context) {
			moon.runMiddleware(i+1, ctx, w, r, end)
		}
		moon.mws[i](ctx, next).ServeHTTP(w, r)
	}
}

// Set the chain final handler
func (moon Moon) Then(handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ctx context.Context
		if Context != nil {
			ctx = Context(r)
		} else {
			ctx = context.TODO()
		}

		moon.runMiddleware(0, ctx, w, r, handler)
	})
}

// Set the chain final handler with a function
func (moon Moon) ThenFunc(fn HandlerFunc) http.Handler {
	handler := func(ctx context.Context) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(ctx, w, r)
		})
	}
	return moon.Then(handler)
}

// Adapt is used to wrap a 3rd party middleware to use with moon.New
// the signature supported is
// func (http.Handler) http.Handler
func Adapt(fn func(http.Handler) http.Handler) Middleware {
	return func(ctx context.Context, next Next) http.Handler {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(ctx)
		})
		return fn(h)
	}
}
