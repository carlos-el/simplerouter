# simplerouter
'simplerouter' is go package that provides a HTTP request router built directly on top of the standard library net/http package. Its main focus is to make a simpler and more expressive API for defining routes and handling HTTP requests, while maintaining full compatibility with net/http.

### Features
- Simple and expressive API for defining routes and handling HTTP requests.
- Full net/http compatibility. Built directly on top of the standard library, the router builds the routes into a net/http `http.ServerMux`. Therefore it works seamlessly with existing net/http handlers and middleware.
- Straightforward middleware integration. Add middleware directly to routes without adding complexity.
- Nested routing support. Allows organizing routes in a hierarchical manner.

### Examples
Examples for route composition patterns and middleware integration can be found in the `_examples` directory.  
Each folder is an independent package that can be run individually.