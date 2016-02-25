package moon

import (
	"net/http"

	"golang.org/x/net/context"
)

type Middleware func(context.Context, Next) http.Handler
type Handler func(context.Context) http.Handler
type Next func(context.Context)
type HandlerWithContext interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request)
}
type HandlerWithContextFunc func(context.Context, http.ResponseWriter, *http.Request)

type Moon struct {
	mws []Middleware
}

var (
	Context func(r *http.Request) context.Context
)

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
		next := func(ctx context.Context) {
			moon.runMiddleware(i+1, ctx, w, r, end)
		}
		moon.mws[i](ctx, next).ServeHTTP(w, r)
	}
}

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

func (moon Moon) ThenFunc(fn HandlerWithContextFunc) http.Handler {
	handler := func(ctx context.Context) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fn(ctx, w, r)
		})
	}
	return moon.Then(handler)
}

func Adapt(fn func(http.Handler) http.Handler) Middleware {
	return func(ctx context.Context, next Next) http.Handler {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next(ctx)
		})
		return fn(h)
	}
}
