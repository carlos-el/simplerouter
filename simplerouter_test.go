package simplerouter_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	r "github.com/carlos-el/simplerouter"
)

func assertCorrect(t testing.TB, got, want any) {
	t.Helper()
	if got != want {
		t.Errorf("got %v want %v", got, want)
	}
}

// middlewareTracker creates a middleware that tracks execution
func middlewareTracker(name string, tracker *[]string) r.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			*tracker = append(*tracker, name)
			next.ServeHTTP(w, req)
		})
	}
}

// handlerWriter creates a handler that writes a specific response
func handlerWriter(response string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(response))
	}
}

func TestNewRoute(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{
			name: "empty path",
			path: "",
		},
		{
			name: "simple api path",
			path: "/api/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.NewRoute(tt.path)

			assertCorrect(t, got.Path, tt.path)
			assertCorrect(t, len(got.Middlewares), 0)
			assertCorrect(t, len(got.Routes), 0)
			assertCorrect(t, got.Method, "")
			// Special handling for Handler comparison since it's a function type
			if got.Handler != nil {
				t.Errorf("got.Handler = %v, want nil", got.Handler)
			}
		})
	}
}

func TestHTTPMethodRoutes(t *testing.T) {
	methods := []struct {
		name       string
		controller func(http.HandlerFunc) *r.Route
	}{
		{name: http.MethodGet, controller: r.Get},
		{name: http.MethodHead, controller: r.Head},
		{name: http.MethodPost, controller: r.Post},
		{name: http.MethodPut, controller: r.Put},
		{name: http.MethodDelete, controller: r.Delete},
		{name: http.MethodPatch, controller: r.Patch},
		{name: http.MethodConnect, controller: r.Connect},
		{name: http.MethodOptions, controller: r.Options},
		{name: http.MethodTrace, controller: r.Trace},
		{name: "", controller: r.All},
	}

	tests := []struct {
		name    string
		handler http.HandlerFunc
	}{
		{
			name:    "correct handler",
			handler: handlerWriter("test"),
		},
		{
			name:    "nil handler",
			handler: nil,
		},
	}

	for _, method := range methods {
		for _, tt := range tests {
			testName := method.name + "/" + tt.name
			t.Run(testName, func(t *testing.T) {
				got := method.controller(tt.handler)

				assertCorrect(t, reflect.ValueOf(got.Handler).Pointer(), reflect.ValueOf(tt.handler).Pointer())
				assertCorrect(t, got.Method, method.name)
				assertCorrect(t, got.Path, "")
				assertCorrect(t, len(got.Middlewares), 0)
				assertCorrect(t, len(got.Routes), 0)
			})
		}
	}
}

// TestAdd tests the Add method functionality with table-driven tests
func TestAdd(t *testing.T) {
	tests := []struct {
		name           string
		initialRoutes  []*r.Route
		routesToAdd    []*r.Route
		expectedLength int
	}{
		{
			name:           "add single route to empty route",
			initialRoutes:  []*r.Route{},
			routesToAdd:    []*r.Route{r.Get(handlerWriter("h1"))},
			expectedLength: 1,
		},
		{
			name:           "add multiple routes at once",
			initialRoutes:  []*r.Route{},
			routesToAdd:    []*r.Route{r.Get(handlerWriter("h1")), r.Post(handlerWriter("h2")), r.Delete(handlerWriter("h1"))},
			expectedLength: 3,
		},
		{
			name:           "add to existing routes",
			initialRoutes:  []*r.Route{r.Get(handlerWriter("h1")), r.Post(handlerWriter("h2"))},
			routesToAdd:    []*r.Route{r.Put(handlerWriter("h1")), r.Patch(handlerWriter("h2"))},
			expectedLength: 4,
		},
		{
			name:           "add no routes (empty variadic)",
			initialRoutes:  []*r.Route{r.Get(handlerWriter("h1"))},
			routesToAdd:    []*r.Route{},
			expectedLength: 1,
		},
		{
			name:           "add nested routes",
			initialRoutes:  []*r.Route{},
			routesToAdd:    []*r.Route{r.NewRoute("/users").Add(r.Get(handlerWriter("h1"))), r.NewRoute("/posts").Add(r.Post(handlerWriter("h2")))},
			expectedLength: 2,
		},
		{
			name:           "add routes with middlewares",
			initialRoutes:  []*r.Route{},
			routesToAdd:    []*r.Route{r.NewRoute("/api").Use(middlewareTracker("mw1", &[]string{})).Add(r.Get(handlerWriter("h1")))},
			expectedLength: 1,
		},
		{
			name:           "add routes with paths",
			initialRoutes:  []*r.Route{},
			routesToAdd:    []*r.Route{r.NewRoute("/users"), r.NewRoute("/posts"), r.NewRoute("/comments")},
			expectedLength: 3,
		},
		{
			name:           "add mixed route types",
			initialRoutes:  []*r.Route{r.NewRoute("/base")},
			routesToAdd:    []*r.Route{r.Get(handlerWriter("h1")), r.NewRoute("/sub"), r.Post(handlerWriter("h2"))},
			expectedLength: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new route for testing
			route := r.NewRoute("/test")

			// Add initial routes if any
			if len(tt.initialRoutes) > 0 {
				route.Add(tt.initialRoutes...)
			}

			// Test the Add method
			result := route.Add(tt.routesToAdd...)

			// Verify the method returns the same route instance (for chaining)
			if result != route {
				t.Errorf("Add() should return the same route instance for method chaining")
			}

			// Verify the correct number of routes
			if len(route.Routes) != tt.expectedLength {
				t.Errorf("Add() resulted in %d routes, want %d", len(route.Routes), tt.expectedLength)
			}

			// Verify that the added routes are in the correct positions
			expectedStartIndex := len(tt.initialRoutes)
			for i, expectedRoute := range tt.routesToAdd {
				actualIndex := expectedStartIndex + i
				if actualIndex < len(route.Routes) {
					if route.Routes[actualIndex] != expectedRoute {
						t.Errorf("Route at index %d is not the expected route", actualIndex)
					}
				}
			}

			// Verify initial routes are still there and in correct order
			for i, expectedRoute := range tt.initialRoutes {
				if i < len(route.Routes) && route.Routes[i] != expectedRoute {
					t.Errorf("Initial route at index %d was modified or moved", i)
				}
			}
		})
	}
}

// TestAddWithNilRoute tests that adding nil routes causes a panic
func TestAddWithNilRoute(t *testing.T) {
	route := r.NewRoute("/test")

	// Adding nil as a route element should panic
	defer func() {
		if recover() == nil {
			t.Error("Expected Add(nil) to panic, but it didn't")
		}
	}()

	route.Add(nil)
}

// TestAddWithMixedNilRoute tests that adding multiple routes where one is nil causes a panic
func TestAddWithMixedNilRoute(t *testing.T) {
	route := r.NewRoute("/test")
	validRoute := r.Get(handlerWriter("h1"))

	// Adding multiple routes where one is nil should panic
	defer func() {
		if recover() == nil {
			t.Error("Expected Add() with nil route to panic, but it didn't")
		}
	}()

	route.Add(validRoute, nil)
}

// TestUse tests the Use method functionality with table-driven tests
func TestUse(t *testing.T) {
	tests := []struct {
		name               string
		initialMiddlewares []r.Middleware
		middlewaresToAdd   []r.Middleware
		expectedLength     int
	}{
		{
			name:               "add single middleware to empty route",
			initialMiddlewares: []r.Middleware{},
			middlewaresToAdd:   []r.Middleware{middlewareTracker("m1", &[]string{})},
			expectedLength:     1,
		},
		{
			name:               "add multiple middlewares at once",
			initialMiddlewares: []r.Middleware{},
			middlewaresToAdd:   []r.Middleware{middlewareTracker("m1", &[]string{}), middlewareTracker("m2", &[]string{})},
			expectedLength:     2,
		},
		{
			name:               "add to existing middlewares",
			initialMiddlewares: []r.Middleware{middlewareTracker("m1", &[]string{})},
			middlewaresToAdd:   []r.Middleware{middlewareTracker("m2", &[]string{})},
			expectedLength:     2,
		},
		{
			name:               "add no middlewares (empty variadic)",
			initialMiddlewares: []r.Middleware{middlewareTracker("m1", &[]string{})},
			middlewaresToAdd:   []r.Middleware{},
			expectedLength:     1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new route for testing
			route := r.NewRoute("/test")

			// Add initial middlewares if any
			if len(tt.initialMiddlewares) > 0 {
				route.Use(tt.initialMiddlewares...)
			}

			// Test the Use method
			result := route.Use(tt.middlewaresToAdd...)

			// Verify the method returns the same route instance (for chaining)
			if result != route {
				t.Errorf("Use() should return the same route instance for method chaining")
			}

			// Verify the correct number of middlewares
			if len(route.Middlewares) != tt.expectedLength {
				t.Errorf("Use() resulted in %d middlewares, want %d", len(route.Middlewares), tt.expectedLength)
			}

			// Verify that the added middlewares are in the correct positions
			expectedStartIndex := len(tt.initialMiddlewares)
			for i, expectedMiddleware := range tt.middlewaresToAdd {
				actualIndex := expectedStartIndex + i
				if actualIndex < len(route.Middlewares) {
					if reflect.ValueOf(route.Middlewares[actualIndex]).Pointer() != reflect.ValueOf(expectedMiddleware).Pointer() {
						t.Errorf("Middleware at index %d is not the expected middleware", actualIndex)
					}
				}
			}

			// Verify initial middlewares are still there and in correct order
			for i, expectedMiddleware := range tt.initialMiddlewares {
				if i < len(route.Middlewares) {
					if reflect.ValueOf(route.Middlewares[i]).Pointer() != reflect.ValueOf(expectedMiddleware).Pointer() {
						t.Errorf("Initial middleware at index %d was modified or moved", i)
					}
				}
			}
		})
	}
}

// TestUseWithNilMiddleware tests that adding nil middlewares causes a panic
func TestUseWithNilMiddleware(t *testing.T) {
	route := r.NewRoute("/test")

	// Adding nil as a middleware element should panic
	defer func() {
		if recover() == nil {
			t.Error("Expected Use(nil) to panic, but it didn't")
		}
	}()

	route.Use(nil)
}

// TestUseWithMixedNilMiddleware tests that adding multiple middlewares where one is nil causes a panic
func TestUseWithMixedNilMiddleware(t *testing.T) {
	route := r.NewRoute("/test")

	// Adding multiple middlewares where one is nil should panic
	defer func() {
		if recover() == nil {
			t.Error("Expected Use() with nil middleware to panic, but it didn't")
		}
	}()

	route.Use(middlewareTracker("m1", &[]string{}), nil)
}

// TestMount tests the Mount function with table-driven tests
func TestMount(t *testing.T) {
	tests := []struct {
		name           string
		setupRoute     func(mwList *[]string) *r.Route
		method         string
		path           string
		expectedMws    []string
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "route not found",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(r.Get(handlerWriter("foo get")))
			},
			method:         "GET",
			path:           "/api/not-found",
			expectedMws:    []string{},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n", // Default http mux not found response
		},
		{
			name: "route handler not allowed",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(r.Get(handlerWriter("foo get")))
			},
			method:         "POST",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method Not Allowed\n", // Default http mux method not allowed response
		},
		{
			name: "root route",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/").Add(
					r.Get(handlerWriter("root")),
				)
			},
			method:         "GET",
			path:           "/",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "root",
		},
		{
			name: "route with multiple method handlers",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Get(handlerWriter("foo get")),
					r.Post(handlerWriter("foo post")),
					r.Patch(handlerWriter("foo patch")),
				)
			},
			method:         "POST",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo post",
		},
		{
			name: "route with All method handler matches concrete handler",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Get(handlerWriter("foo get")),
					r.All(handlerWriter("foo all")),
				)
			},
			method:         "GET",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get",
		},
		{
			name: "all http methods, get",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Get(handlerWriter("foo get")),
				)
			},
			method:         "GET",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get",
		},
		{
			name: "all http methods, head",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Head(handlerWriter("foo head")),
				)
			},
			method:         "HEAD",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo head",
		},
		{
			name: "all http methods, post",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Post(handlerWriter("foo post")),
				)
			},
			method:         "POST",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo post",
		},
		{
			name: "all http methods, put",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Put(handlerWriter("foo put")),
				)
			},
			method:         "PUT",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo put",
		},
		{
			name: "all http methods, patch",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Patch(handlerWriter("foo patch")),
				)
			},
			method:         "PATCH",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo patch",
		},
		{
			name: "all http methods, delete",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Delete(handlerWriter("foo delete")),
				)
			},
			method:         "DELETE",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo delete",
		},
		{
			name: "all http methods, connect",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Connect(handlerWriter("foo connect")),
				)
			},
			method:         "CONNECT",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo connect",
		},
		{
			name: "all http methods, options",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Options(handlerWriter("foo options")),
				)
			},
			method:         "OPTIONS",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo options",
		},
		{
			name: "all http methods, trace",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Add(
					r.Trace(handlerWriter("foo trace")),
				)
			},
			method:         "TRACE",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo trace",
		},
		{
			name: "nested routes",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api").Add(
					r.NewRoute("/foo").Add(
						r.Get(handlerWriter("foo get")),
					),
					r.NewRoute("/bar").Add(
						r.Get(handlerWriter("bar get")),
					),
				)
			},
			method:         "GET",
			path:           "/api/foo",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get",
		},
		{
			name: "parametrized nested route",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api").Add(
					r.NewRoute("/foo/{id}").Add(
						r.Get(handlerWriter("foo get id")),
					),
				)
			},
			method:         "GET",
			path:           "/api/foo/123",
			expectedMws:    []string{},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get id",
		},
		{
			name: "middleware execution order",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Use(
					middlewareTracker("mw1", mwTrackerSlice),
					middlewareTracker("mw2", mwTrackerSlice),
				).Add(
					r.Get(handlerWriter("foo get")),
				)
			},
			method:         "GET",
			path:           "/api/foo",
			expectedMws:    []string{"mw1", "mw2"},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get",
		},
		{
			name: "nested middleware execution order",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Use(
					middlewareTracker("mw1", mwTrackerSlice),
				).Add(
					r.Get(handlerWriter("foo get")).Use(
						middlewareTracker("mw3", mwTrackerSlice),
					),
				).Use(
					middlewareTracker("mw2", mwTrackerSlice),
				)
			},
			method:         "GET",
			path:           "/api/foo",
			expectedMws:    []string{"mw1", "mw2", "mw3"},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo get",
		},
		{
			name: "nested middleware",
			setupRoute: func(mwTrackerSlice *[]string) *r.Route {
				return r.NewRoute("/api/foo").Use(
					middlewareTracker("mw1", mwTrackerSlice),
				).Add(
					r.NewRoute("/bar").Use(
						middlewareTracker("mw2", mwTrackerSlice),
					).Add(
						r.Get(handlerWriter("foo bar get")).Use(
							middlewareTracker("mw3", mwTrackerSlice),
						),
					),
				)
			},
			method:         "GET",
			path:           "/api/foo/bar",
			expectedMws:    []string{"mw1", "mw2", "mw3"},
			expectedStatus: http.StatusOK,
			expectedBody:   "foo bar get",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a fresh middleware tracking list for each test
			mwTrackerSlice := []string{}
			route := tt.setupRoute(&mwTrackerSlice)
			mux := route.Mount()

			// Verify mux is not nil
			if mux == nil {
				t.Fatal("Mount() returned nil ServeMux")
			}

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			// Serve the request
			mux.ServeHTTP(w, req)

			// Verify response status
			if w.Code != tt.expectedStatus {
				t.Errorf("Status = %d, want %d", w.Code, tt.expectedStatus)
			}

			// Verify response body
			if w.Body.String() != tt.expectedBody {
				t.Errorf("Body = %q, want %q", w.Body.String(), tt.expectedBody)
			}

			// Verify middleware execution order
			if !reflect.DeepEqual(mwTrackerSlice, tt.expectedMws) {
				t.Errorf("Middlewares executed = %v, want %v", mwTrackerSlice, tt.expectedMws)
			}
		})
	}
}
