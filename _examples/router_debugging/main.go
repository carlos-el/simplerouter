// This example demonstrates how debug the router structure by using the walk functionality.
//
// Run the server from this directory using the following command:
//
//	$ go run main.go
//
// Check the console output to see the structure of the mounted routes and the middlewares and handlers associated.
// Execute the endpoints using any http client to check the routing and middleware execution order.
package main

import (
	"fmt"
	"net/http"
	"reflect"
	"runtime"

	r "github.com/carlos-el/simplerouter"
)

// Dummy middlewares
func generalMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GeneralMiddleware")
		next.ServeHTTP(w, r)
	})
}

func fooMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("FooMiddleware")
		next.ServeHTTP(w, r)
	})
}

func getFooMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("GetFooMiddleware")
		next.ServeHTTP(w, r)
	})
}

func barMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("BarMiddleware")
		next.ServeHTTP(w, r)
	})
}

// Dummy handlers
func getFooHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GetFooHandler")
	w.Write([]byte("getFooHandler"))
}

func postBarHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("PostBarHandler")
	w.Write([]byte("postBarHandler"))
}

func main() {
	// Walker function for describing the final working endpoints structure.
	// Describes the middlewares that apply and the handler functions used in the right order
	var walker = func(route *r.Route, path string, middlewares []r.Middleware) {
		if route.Handler != nil {
			fmt.Println(path + route.Path + " " + route.Method)
			for _, mw := range append(middlewares, route.Middlewares...) {
				fmt.Println("\t" + runtime.FuncForPC(reflect.ValueOf(mw).Pointer()).Name())
			}
			fmt.Println("\t" + runtime.FuncForPC(reflect.ValueOf(route.Handler).Pointer()).Name())
		}
	}

	router := r.NewRoute("/api").Use(
		generalMiddleware,
	).Add(
		r.NewRoute("/foo").Use(fooMiddleware).Add(
			r.Get(getFooHandler).Use(getFooMiddleware),
		),
		r.NewRoute("/bar").Use(barMiddleware).Add(
			r.Post(postBarHandler),
		),
	).MountAndWalk(walker)
	// Console output:
	// /api/foo GET
	// 	main.generalMiddleware
	// 	main.fooMiddleware
	// 	main.getFooMiddleware
	// 	main.getFooHandler
	// /api/bar POST
	// 	main.generalMiddleware
	// 	main.barMiddleware
	// 	main.postBarHandler

	http.ListenAndServe(":5000", router)
}
