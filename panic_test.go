package moon

import (
	"net/http"
	"testing"

	"golang.org/x/net/context"
)

func throwPanicHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("No way")
	})
}

func TestPanic(t *testing.T) {
	mws := New(Panic).Then(throwPanicHandler)
	res := serveAndRequest(mws, false)
	assertHasPrefix(t, res, "Panic:\nNo way\n\n")
}
