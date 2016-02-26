package moon

import (
	"fmt"
	"net/http"
	"runtime"

	"golang.org/x/net/context"
)

// Middlware Panic, catch the panic throw from the inner middlewares
// Write http.StatusInternalServerError and the goroutine stacktrace
func Panic(ctx context.Context, next Next) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if r := recover(); r != nil {
				w.WriteHeader(http.StatusInternalServerError)
				stack := make([]byte, 4096)
				runtime.Stack(stack, false)
				fmt.Fprintf(w, "Panic:\n%v\n\n%v", r, string(stack))
			}
		}()
		next(ctx)
	})
}
