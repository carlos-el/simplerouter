// Package simplerouter provides a minimalistic net/http
// compatible router focused on expressiveness and simplicity.
package simplerouter

import (
	"net/http"
)

// Middleware is any function that takes an http.Handler and returns an http.Handler.
type Middleware func(http.Handler) http.Handler

// applyMiddleware applies a list of middlewares to the next http.Handler.
// It returns a new http.Handler that applies the middlewares in reverse order.
func applyMiddleware(mws ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(mws) - 1; i >= 0; i-- {
			x := mws[i]
			next = x(next)
		}

		return next
	}
}

// Route represents a route in the router.
// It stores the information about the route's path, middlewares, child routes, handler, and HTTP method.
type Route struct {
	Path        string
	Middlewares []Middleware
	Routes      []*Route
	Handler     http.HandlerFunc
	Method      string
}

// NewRoute creates a new Route with the given path path.
// It initializes the route with an empty list of middlewares and child routes.
func NewRoute(path string) *Route {
	return &Route{
		Path:        path,
		Middlewares: []Middleware{},
		Routes:      []*Route{},
		Handler:     nil,
		Method:      "",
	}
}

// Use adds middlewares that execute before the route's handlers or child routes.
func (r *Route) Use(middlewares ...Middleware) *Route {
	for _, mw := range middlewares {
		if mw == nil {
			panic("middlewares parameter cannot contain nil middlewares")
		}
	}
	r.Middlewares = append(r.Middlewares, middlewares...)
	return r
}

// Add adds child routes to the current route.
func (r *Route) Add(routes ...*Route) *Route {
	for _, route := range routes {
		if route == nil {
			panic("routes parameter cannot contain nil routes")
		}
	}
	r.Routes = append(r.Routes, routes...)
	return r
}

// Returns a Route with the handler associated to the GET http method and no path.
func Get(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodGet}
}

// Returns a Route with the handler associated to the HEAD http method and no path.
func Head(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodHead}
}

// Returns a Route with the handler associated to the POST http method and no path.
func Post(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodPost}
}

// Returns a Route with the handler associated to the PUT http method and no path.
func Put(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodPut}
}

// Returns a Route with the handler associated to the PATCH http method and no path.
func Patch(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodPatch}
}

// Returns a Route with the handler associated to the DELETE http method and no path.
func Delete(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodDelete}
}

// Returns a Route with the handler associated to the CONNECT http method and no path.
func Connect(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodConnect}
}

// Returns a Route with the handler associated to the OPTIONS http method and no path.
func Options(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodOptions}
}

// Returns a Route with the handler associated to the TRACE http method and no path.
func Trace(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: http.MethodTrace}
}

// Returns a Route with the handler associated and no path or method.
// This can be used to create a route that matches all methods not explicitly defined (as per the standard lib behavior).
func All(handler http.HandlerFunc) *Route {
	return &Route{Handler: handler, Method: ""}
}

// inspectRoute recursively inspects the route provided and its child routes.
// It applies the paths, middlewares and handlers to the provided http.ServeMux router.
// If a WalkFn is provided, it will be called for each route inspected.
func (r *Route) inspectRoute(
	path string,
	middlewares []Middleware,
	router *http.ServeMux,
	walkFn WalkFn,
) {
	chainedPath := path + r.Path
	chainedMiddleware := append(middlewares, r.Middlewares...)

	if walkFn != nil {
		walkFn(r, path, middlewares)
	}

	if r.Handler != nil {
		router.Handle(
			r.Method+" "+chainedPath,
			applyMiddleware(chainedMiddleware...)(r.Handler),
		)
	}

	for _, route := range r.Routes {
		route.inspectRoute(
			chainedPath,
			chainedMiddleware,
			router,
			walkFn,
		)
	}
}

// Mount returns an http.ServeMux with all the routes and handlers registered.
// Dynamically editing the route after mounting it will not affect the returned http.ServeMux.
// Mounting the route will not validate the route's structure or the presence of handlers.
// It is the user's responsibility to ensure that the route is correctly configured before mounting.
func (r *Route) Mount() *http.ServeMux {
	router := http.NewServeMux()
	r.inspectRoute("", []Middleware{}, router, nil)
	return router
}

// WalkFn is a function type that can be used to walk through the routes as they are mounted.
// It receives the current route and the path path and middlewares of the parent route.
// It can be used for debugging or testing purposes.
type WalkFn func(router *Route, path string, middlewares []Middleware)

// MountAndWalk does the same as [Route.Mount], but requires a WalkFn to be provided.
// The WalkFn will be called for each route and subroute,
// allowing for custom debugging or logging of the routes.
func (r *Route) MountAndWalk(walkFn WalkFn) *http.ServeMux {
	if walkFn == nil {
		panic("walkFn parameter cannot be nil")
	}

	router := http.NewServeMux()
	r.inspectRoute("", []Middleware{}, router, walkFn)
	return router
}
