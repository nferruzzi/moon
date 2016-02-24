# Moon
[![Build Status](https://travis-ci.org/nferruzzi/moon.svg?branch=master)](https://travis-ci.org/nferruzzi/moon)
[![Coverage Status](https://coveralls.io/repos/github/nferruzzi/moon/badge.svg?branch=master)](https://coveralls.io/github/nferruzzi/moon?branch=master)
[![MIT licensed](https://img.shields.io/badge/license-MIT-blue.svg)](./LICENSE)

Moon is a simple middleware chaining system with requests context handled by [/x/net/context](https://blog.golang.org/context) context.Context

No magic GO code is required to use Moon.

### Quick tour

Middlewares use this signature

```go
func (context.Context, moon.HandlerWithContext) http.Handler
```

and final handler

```go
func handler(context.Context) http.Handler
```

Middlewares are chained with `Moon.New`

```go
middlewares := moon.New(middleware1, middleware2, middleware3, ...)
```

`moon.New` returns a `Moon` struct.

The final handler is appended by passing `moon.Handler` to `Moon.Then`

```go
r.Handle("/api", middlewares.Then(handler))
```

or by passing a function to `Moon.ThenFunc`

```go
r.Handle("/api", middlewares.ThenFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
  ...
}))
```

Inside a middleware you can advance the chain by calling `next.ServeHTTP(ctx, w, r)`

```go
func Middleware(ctx context.Context, next moon.HandlerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // ...
    next.ServeHTTP(ctx, w, r)
    // ...
	})
}
```

### contex.TODO() or appengine.NewContext(request) ?

By default a `context.Context` is created for each request and is an instance of

```go
ctx := context.TODO()
```

this behaviour can be modified by setting a callback for `moon.Context`
ie. if you run you app in appengine:

```go
func init() {
  ...

  // just setup it once; then this function is called for every request
  moon.Context = func(r *http.Request) {
    return appengine.NewContext(r)
  }  

  ...
}
```


### 3rd party middlewares

Compatibility is provided with all 3rd party middlewares using the following signature

```go
func (http.Handler) http.Handler
```

just wrap the function with

`moon.Adapt`

ie. GOJI SimpleBasicAuth

```go
goji_middleware := moon.Adapt(httpauth.SimpleBasicAuth("user", "pass"))
middlewares := moon.New(goji_middlware, ...).Then(...)

```

### Quick example

```go
// Appengine test app
package app

import (
	"fmt"
	"net/http"

	"google.golang.org/appengine"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"github.com/nferruzzi/moon"
)

func MWRequireJSON(ctx context.Context, next moon.HandlerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ct := r.Header.Get("Content-Type")
		if ct == "application/json" {
			ctx = context.WithValue(ctx, "Content-Type", ct)
			next.ServeHTTP(ctx, w, r)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	})
}

func MWRequireUser(ctx context.Context, next moon.HandlerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// let's pretend some checks are made
		ctx = context.WithValue(ctx, "User", "user")
		next.ServeHTTP(ctx, w, r)
	})
}

func handler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello %v! thanks for your %v", ctx.Value("User"), ctx.Value("Content-Type"))
	})
}

func init() {
	// Appengine only: configure moon to get the context from the request
	// Common default is context.TODO()
	moon.Context = func(r *http.Request) context.Context {
		return appengine.NewContext(r)
	}

	// Middlewares chain
	middlewares := moon.New(MWRequireJSON, MWRequireUser)

	// Use gorilla mux to route and handle the request
	r := mux.NewRouter()
	r.Handle("/api", middlewares.Then(handler))

	http.Handle("/", r)
}

// curl -H Content-Type:application/json http://localhost:8080/api
// Hello user! thanks for your application/json✔ ~
```

### Stability

API is under development.

### Acknowledgements

Thanks to the authors of the package below, I got a lot of idea from your code

Using contex.Context inspired by kami and appengine

https://github.com/guregu/kami

Middleware chaining inspired by Alice

https://github.com/justinas/alice

and by Stack

https://github.com/alexedwards/stack

Bidirectional middleware flow inspired by Negroni

https://github.com/codegangsta/negroni

Tested with GOJI middlewares

https://github.com/goji/goji
