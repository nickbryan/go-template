# Go Template
One of the biggest problems developers face when starting with Go is where do I put things and what should I use? This
project aims to solve that by creating a foundation on which future Go based micro-services can be created. The Go community
tends to recommend that we do not use a heavy MVC framework for our micro-services and apis. Go has an incredible standard
library which we can get most of the functionality we need from, however there are some packages out there that make it
even easier to work with certain parts of an application or add extra flexibility.

This project is intended to be a starting point for a new Go based micro-service. The reason we have not abstracted this
into a library is that we want developers to be able to change and modify any part of the application as they become 
more confident with Go and learn new patterns and practices.

* [Goals](#goals)
* [Packages chosen](#packages-chosen)
* [Getting started](#getting-started)
    * [Project structure](#project-structure)
    * [HTTP](#http)
        * [Server](#server)
        * [Handler](#handler)      
* [Build and deployment](#build-and-deployment)
    * [Development](#development)
    * [Production](#production)
* [Current pain points](#current-pain-points)

## Goals
The starter aims to meet the following goals:
* Familiar application structure that is easy to navigate.
* Easy configuration.
* Preconfigured deployment pipeline so that developers can start pushing to production straight away.
* High cohesion/loose coupling.
* Integration tests that can be run against a real database.
* Fast test suites through parallelisation.
* Developer friendly.

## Packages chosen
The following set of packages provide us with a solid foundation to build our applications upon. Most Go api frameworks 
only provide what `gorilla/mux` is doing unless we were to go with a fully fledged MVC framework but that's not really what 
Go is about.

This provides us with configuration, a database driver for postgres, a query builder, migrations, easy struct scanning of database 
records, routing, validation, structured logging, test assertions and integration tests against the database.

* [air](https://github.com/cosmtrek/air) - ☁️ Live reload for Go apps
* [cobra](https://github.com/spf13/cobra) - Cobra is both a library for creating powerful modern CLI applications as well as a program to generate applications and command files.
* [dockertest](https://github.com/ory/dockertest) - Use Docker to run your Go language integration tests against third party services.
* [Gabs](https://github.com/Jeffail/gabs) - Gabs is a small utility for dealing with dynamic or unknown JSON structures in Go.
* [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations. CLI and Golang library.
* [gorilla/mux](https://github.com/gorilla/mux) - Package gorilla/mux implements a request router and dispatcher for matching incoming requests to their respective handler.
* [ozzo-validation](https://github.com/go-ozzo/ozzo-validation) - ozzo-validation is a Go package that provides configurable and extensible data validation capabilities.
* [pgx](https://github.com/JackC/pgx) - pgx is a pure Go driver and toolkit for PostgreSQL.
* [scany](https://github.com/georgysavva/scany) - Scany allows developers to scan complex data from a database into Go structs.
* [Squirrel](https://github.com/Masterminds/squirrel) - Squirrel helps you build SQL queries from composable parts.
* [testify](https://github.com/stretchr/testify) - Set of packages that provide many tools for testifying that your code will behave as you intend.
* [uuid](https://github.com/google/uuid) - The uuid package generates and inspects UUIDs
* [viper](https://github.com/spf13/viper) - Viper is a complete configuration solution for Go applications including 12-Factor apps.
* [zap](https://github.com/uber-go/zap) - Blazing fast, structured, leveled logging in Go.

## Getting started
The following is intended to give a rough guide to getting up and running with the starter template.

### Project structure
The following gives a brief overview of the project layout and what each package/file is intended for. This is not a strict
layout and is only intended as a guide for developers starting a new project. Packages should have meaningful names and
encapsulate one thing well.

```text
deploy   <-- This is where all Docker, Terraform and Helm configurations live.
service   <-- This is where the main application service code lives.
├── app   <-- The app package registers the main application services that 
|         |   form either the DefaultEnvironment or TestEnvironment for the app.
│         ├── config.go   <-- Loads the application config from the filesystem into the Environment.
│         ├── environment.go   <-- Key application services are registered and exported here.
│         ├── migrations   <-- Database migrations, these are embedded into the binary at compile time.
│         │         ├── 20210220170406_create_customers_table.down.sql
│         │         └── 20210220170406_create_customers_table.up.sql
│         ├── postgres.go   <-- All database initialisation code is here. We expose the 
|         |                     connection pool and query builder to the Environment.
│         └── validator.go   <-- Helper function for validation, wraps errors.
├── cmd   <-- Command line entry points to the application live here.
│         ├── root.go   <-- This is required by cobra to initialise the main terminal command for the app.
│         └── server.go   <-- This is the command that we will run to start the HTTP server and serve the handlers.
├── config.yaml   <-- Application configuration can be registered here.
├── config_test.yaml   <-- The above config can be overridden for tests.
├── domain   <-- The home of all our business logic.
│         └── customer   <-- Meaningful package names to fit the domain concepts.
│             └── customer.go   <-- Repository interface, Entity and functions for dealing with a customer.
├── go.mod   <-- App dependencies, similar to composer.json or package.json.
├── go.sum   <-- Dependency lock file created by go mod.
├── infrastructure   <-- All third party integrations should be declared here.
│         └── postgres   <-- All postgres related code in this package.
│             ├── customer.go   <-- Implements the customer.Repository from the domain/customer package.
│             └── postgrestest   <-- Test helpers for interacting with postgres database. Such as AssertDatabaseHas
|                 |                  which checks if a record exists in a given table.
│                 └── assertions.go
├── main.go   <-- Calls into cobra to initialise the command line interface.
├── test.env   <-- Environment variables can be overriden for tests here.
└── transport   <-- Our main entry points into the application logic. Currently there is only rest but in the future
    |               we could have amqp, grpc, sqs etc.
    └── rest   <-- All code for handling rest requests goes in here.
        ├── auth.go   <-- Middleware for handling authentication via JWT.
        ├── customers   <-- Handlers for the customer resource. Can be thought of as controllers or actions.
        │         ├── create_handler.go
        │         └── create_handler_test.go   <-- Integration tests for the handlers. These interact with the database.
        ├── handler.go   <-- The Handler definition for all rest requests.
        ├── health   <-- Handler for the health check endpoint.
        │         ├── check_handler.go
        │         └── check_handler_test.go
        ├── resttest   <-- Test helpers for making http rest requests with JSON.
        │         └── server.go
        ├── server.go   <-- Wraps the http.Server to allow for graceful shutdown and configuration. The router is created
        |                   here.
        ├── server_test.go
        ├── util.go   <-- Helpers for dealing with JSON requests and responses.
        └── util_test.go
```

### HTTP
I mainly write JSON based REST apis so this starter template has been designed with that in mind. The 
`transport/rest` package encapsulates all the code related to the HTTP layer of our application. At the root, you can find
the code for setting up the `Server`, the main `Handler` definition, helpers for dealing with requests and responses and 
any global middleware required by the handlers. The sub packages are intended to hold all `Handler`s for dealing with a 
specific resource (actions/controllers).

#### Server
The code for the `Server` can be found in `transport/rest/server.go` and should be well documented if you want to have a
look at what is going in there. The `Server` wraps and starts the standard libraries `http.Server`. We do this so that we can
register a router as the servers main handler and allow a graceful shutdown when we receive an interrupt signal. The server
is started in `cmd/server.go`.

#### Handler
The handler definition can be found in `transport/rest/handler.go`. A handler is registered with the `Server` by calling:
```go
s := rest.NewServer(...)

s.RegisterHandlers(
	customers.NewIndexHandler(/* Inject dependencies here... */),
	customers.NewCreateHandler(/* Inject dependencies here... */),
	customers.NewUpdateHandler(/* Inject dependencies here... */),
)

s.Start()
```

The handler is defined as follows:
```go
package rest

// Handler is responsible for defining a HTTP request route and corresponding handler.
type Handler struct {
	// Route receives a route to modify, like adding path, methods, etc.
	Route func(r *mux.Route)

	// Middleware allows wrapping the Func in middleware handlers.
	Middleware func(next http.Handler) http.HandlerFunc

	// Func will be registered with the router.
	Func http.HandlerFunc
}
```

We call a function to create our handler which allows us to form a closured environment where we can do some setup work,
take in our dependencies and return the `Handler`.

A simple example of a handler can be seen below:
```go
func NewCreateHandler(repo customer.Repository) rest.Handler {
	// We can defined our request and response structs here. These
	// will allow us to unmarshall JSON from the request and marshal 
	// to the response objects.
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	
	type response struct {
		ID string `json:"id"`	
	}
    
	// Anything we do before the return will only be called/created
	// once when we register the Handler with the Server.

	return rest.Handler{
		Route: func(r *mux.Route) {
			// Here we can declare our route path and the methods it
			// should respond to.
			r.Path("/customers").Methods(http.MethodPost)
		},
		Middleware: func(next http.Handler) http.HandlerFunc {
			// next will be the Func that is declared below.
			// We can wrap it in any middleware here.
			return rest.RequireJWTAuthentication(next)
		}
		Func: func(w http.ResponseWriter, r *http.Request) {
			// This is the main function for our http requests. This will
			// be called on each request.
			
			// Declare our variable to decode the request to.
			var req request

			// rest.Decode is a helper for unmarshalling the request payload
			// into a request struct as defined above.
			if err := rest.Decode(r, &req); err != nil {
				rest.RespondError(w, http.StatusBadRequest, err)

				return
			}

			// We can run any validation we require against the request struct
			// defined above by passing it as a reference to the `app.Validate`
			// helper along with the validation field definitions.
			if errs := app.Validate(&req,
				validation.Field(&req.Username, validation.Required, is.Email, usernameUniqueRule{repo, r.Context()}),
				validation.Field(&req.Password, validation.Required, validation.Length(minPassLen, maxPassLen)),
			); errs != nil {
				// This helper will format the errors properly and write them
				// to the http.ResponseWriter.
				rest.RespondValidationFailed(w, errs)

				return
			}

			// Create our Customer object.
			cust, err := customer.New(req.Username, req.Password)
			if err != nil && errors.Is(err, customer.ErrGeneratePassword) {
				// TODO: handle error
			}

			// Add it to the database through the repository.
			if err := repo.Add(r.Context(), cust); err != nil {
				// TODO: handle error
			}

			w.WriteHeader(http.StatusCreated)
		},
	}
}
```

## Build and deployment
### Development
The project is currently setup for use with a postgres database. We use Docker Compose to create a container for postgres and
a container for [air](https://github.com/cosmtrek/air) which allows live reloading of the app. This should save some time
having to recompile on every change. The development build also pulls in [golang-migrate](https://github.com/golang-migrate/migrate)
so that we can easily run and create migrations within the docker container.

Unfortunately, because the integration tests currently require that postgres containers are started via Docker, the 
tests must be run locally which means developers are required to have Go installed locally.

The application currently has the following make commands:
```text
Usage: make <target>

The following targets are available:

Local-development
  docker-up   <-- Start the docker compose services running in the background.
  docker-down   <-- Stop running containers.
  docker-logs   <-- View application logs.

Migrations
  migrate   <-- Send a custom command to golang-migrate.
  migrate-up   <-- Migrate to the latest migration version.
  migrate-down   <-- Tear down the database to its original state.
  migrate-fresh   <-- Runs migrate-up and migrate-down for a clean database.
  migration   <-- Create a new versioned up and down migration.

Test
  test   <-- Run all of the tests.

Options
  help
```

### Production
...

## Current pain points
* The database layer can be verbose. The query builder helps for dynamic queries, and the exposed connection allows us
to execute SQL direct but this comes with some boiler plate for error handling and mapping.
* Integration tests require docker containers to be created. As it currently stands, the tests must be run on the host machine
requiring the developer to have Go installed locally. (You would probably have this for your IDE anyway.)
* Some parts of the app use the global logger which is bad for parallelism in tests.

