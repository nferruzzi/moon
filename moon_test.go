package moon

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goji/httpauth"
	"golang.org/x/net/context"
)

func assertEquals(t *testing.T, e interface{}, o interface{}) {
	if e != o {
		t.Errorf("\n...expected = %v\n...obtained = %v", e, o)
	}
}

func serveAndRequest(h http.Handler, auth bool) string {
	//auth := base64.StdEncoding.EncodeToString([]byte("user:pass"))
	//r.Header.Set("Authorization", "Basic "+auth)

	ts := httptest.NewServer(h)
	defer ts.Close()

	req, _ := http.NewRequest("GET", ts.URL, nil)

	if auth {
		req.SetBasicAuth("user", "pass")
	}

	res, err := http.DefaultClient.Do(req)
	//res, err := http.Client.Do(req)

	// res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}
	resBody, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//	log.Print(resBody)
	return string(resBody)
}

func tokenMiddlewareA(ctx context.Context, next HandlerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, "tokenA", "123")
		next.ServeHTTP(ctx, w, r)
	})
}

func tokenMiddlewareB(ctx context.Context, next HandlerWithContext) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = context.WithValue(ctx, "tokenB", "456")
		next.ServeHTTP(ctx, w, r)
	})
}

func tokenHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Tokens are: %v, %v", ctx.Value("tokenA"), ctx.Value("tokenB"))
	})
}

func TestContext(t *testing.T) {
	st := New(tokenMiddlewareA, tokenMiddlewareB).Then(tokenHandler)
	res := serveAndRequest(st, false)
	assertEquals(t, "Tokens are: 123, 456", res)
}

func TestContextRoot(t *testing.T) {
	Context = func(r *http.Request) context.Context {
		ctx := context.TODO()
		return context.WithValue(ctx, "tokenA", "789")
	}
	st := New(tokenMiddlewareB).Then(tokenHandler)
	res := serveAndRequest(st, false)
	assertEquals(t, "Tokens are: 789, 456", res)
}

func authHandler(ctx context.Context) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, _ := r.BasicAuth()
		fmt.Fprintf(w, "Hello: %v, %v", u, p)
	})
}

func TestGojiBasicAuthUnauthorized(t *testing.T) {
	st := New(Adapt(httpauth.SimpleBasicAuth("user", "pass"))).Then(authHandler)
	res := serveAndRequest(st, false)
	assertEquals(t, "Unauthorized\n", res)
}

func TestGojiBasicAuthAuthorized(t *testing.T) {
	st := New(Adapt(httpauth.SimpleBasicAuth("user", "pass"))).Then(authHandler)
	res := serveAndRequest(st, true)
	assertEquals(t, "Hello: user, pass", res)
}
