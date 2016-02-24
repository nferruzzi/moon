package moon

import (
	"net/http"

	"golang.org/x/net/context"
)

type HandlerWithContext interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}

type HandlerWithContextFunc func(context.Context, http.ResponseWriter, *http.Request)

func (f HandlerWithContextFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	f(ctx, w, r)
}

type Middleware func(context.Context, HandlerWithContext) http.Handler
type Handler func(context.Context) http.Handler

type Moon struct {
	mws []Middleware
}

var Context func(r *http.Request) context.Context

func New(middlewares ...Middleware) Moon {
	return Moon{mws: middlewares}
}

func (moon Moon) runMiddleware(i int, ctx context.Context, w http.ResponseWriter, r *http.Request, end Handler) {
	if i == len(moon.mws) {
		// end recursion
		if end != nil {
			end(ctx).ServeHTTP(w, r)
		}
	} else {
		// configure call to next middleware
		next := HandlerWithContextFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
			moon.runMiddleware(i+1, ctx, w, r, end)
		})
		moon.mws[i](ctx, next).ServeHTTP(w, r)
	}
}

func (moon Moon) Then(handler Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ctx := appengine.NewContext(r)

		var ctx context.Context
		if Context != nil {
			ctx = Context(r)
		} else {
			ctx = context.TODO()
		}

		moon.runMiddleware(0, ctx, w, r, handler)
	})
}

func Adapt(fn func(http.Handler) http.Handler) Middleware {
	return func(ctx context.Context, next HandlerWithContext) http.Handler {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(ctx, w, r)
		})
		return fn(h)
	}
}
