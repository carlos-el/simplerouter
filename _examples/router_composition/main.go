// This example demonstrates some examples on how to compose routers.
//
// Run the server from this directory using the following command:
//
//	$ go run main.go
//
// Execute the endpoints using any http client to check the routing and middleware execution order.
package main

import (
	"fmt"
	"net/http"

	r "github.com/carlos-el/simplerouter"
)

func createMiddleware(text string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Executing", text)
			next.ServeHTTP(w, r)
		})
	}
}

func createHandler(text string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Executing", text)
		w.Write([]byte(text))
	}
}

func main() {
	// General procedure for composing a router:
	// 1. Create a main route.
	// 2. Add middleware using the Use method.
	// 3. Add child routes using the Add method.
	// 4. Add handlers using the HTTP method functions (Get, Post, All, etc).
	// 5. Finally, call the Mount method to create a net/http http.ServeMux.

	// Simple composition of handlers and middlewares
	fooSubroute := r.NewRoute("").Use(createMiddleware("FooMiddleware")).Add(
		r.NewRoute("/foo").Add(r.Get(createHandler("FooHandler"))),
		r.NewRoute("/foo/qux").Add(r.Get(createHandler("QuxHandler"))),
		r.NewRoute("/foo/qux/:quxId").Add(r.Get(createHandler("QuxHandler by ID"))),
	)

	// Define child routes by adding sub-paths
	barSubroute := r.NewRoute("/bar").Use(createMiddleware("BarMiddleware")).Add(
		r.NewRoute("/baz").Add(
			r.Get(createHandler("GetBazHandler")),
			r.Post(createHandler("PostBazHandler")),
		),
		r.NewRoute("/baz/:bazId").Add(
			r.Patch(createHandler("GetBazHandler by ID")),
			r.Delete(createHandler("PostBazHandler by ID")),
		),
	)

	// Create routers by appending of other routes
	router := r.NewRoute("").Use(
		createMiddleware("GlobalMiddleware1"),
		createMiddleware("GlobalMiddleware2"),
	).Add(
		r.NewRoute("/api/v1").Use(createMiddleware("APIv1Middleware")).Add(
			fooSubroute,
			barSubroute,
			r.NewRoute("/foobar").Add(r.All(createHandler("FooBarHandler"))),
		),
	).Mount()

	http.ListenAndServe(":5000", router)
}
