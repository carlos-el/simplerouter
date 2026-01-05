// This example demonstrates how middleware works.
//
// Run the server from this directory using the following command:
//
//	$ go run main.go
//
// Execute the endpoints using any http client to check the routing and middleware execution order.
package main

import (
	"log"
	"net/http"

	r "github.com/carlos-el/simplerouter"

	"github.com/go-chi/chi/middleware"
)

func firstMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("First middleware")
		next.ServeHTTP(w, r)
	})
}

func secondMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Second middleware")
		next.ServeHTTP(w, r)
	})
}

func thirdMiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Third middleware logger start")
		next.ServeHTTP(w, r)
		log.Printf("Third middleware logger data: [%s] [%s]\n", r.Method, r.URL.Path)
	})
}

func main() {
	router := r.NewRoute("/api/foo").Use(
		// Simple middleware chaining
		firstMiddleware,
		secondMiddleware,
		thirdMiddlewareLogger,
		// Works with external net/http compatible middleware from other popular libraries
		middleware.Logger,
	).Add(
		r.Get(func(w http.ResponseWriter, r *http.Request) {
			log.Println("Handler getFoo")
			w.Write([]byte("Handler getFoo"))
		}),
	).Mount()

	http.ListenAndServe(":5000", router)
}
