# swagger

swagger is [go-swagger](https://github.com/go-swagger/go-swagger) with custom templates.

In Stratoscale, we really like the idea of API-first services, and we also really like Go.
We saw the go-swagger library, and thought that most of it can really help us. Generating code from
swagger files is a big problem with a lot of corner cases, and go-swagger is doing great job other there.

The one thing that we felt missing, is customization of the server to run with our design principles:

* Custom `main` function
* Dependency injection
* Limited scopes with unit testing.

Also:

* Adding you functions to the generated `configure_swagger_*.go` seems to be a burden.
* Lack of Interface that the service implement.
* Complicated and custom http clients and runtime.

Luckily, the go-swagger project gives an option to customize the templates it is using to generate the code.
We re-wrote some of go-swagger's templates, and now we have:

## Server

### The new `restapi` package exposes interfaces

* Those interfaces can implemented by the developer and are the business logic of the service.
* The implementation of those is extensible.
* The implementation is separated from the generated code.

### The `restapi` returns an `http.Handler`

The `restapi.Handler` (see [example](./example/restapi/configure_swagger_petstore.go)) function returns
a standard `http.Handler`

* Given objects that implements the business logic, we can create a simple http handler.
* This handler is standard go http.Handler, so we can now use any other middleware, library, or framework
  that support it.
* This handler is standard, so we understand it better.

## Client

* The new client package exposes interfaces, so functions in our code can receive those
  interfaces which can be mocked for testing.
* The new client has a config that gets an `*url.URL` to customize the endpoint.
* The new client has a config that gets an `http.RoundTripper` to customize client with libraries, middleware or
  frameworks that support the standard library's objects.