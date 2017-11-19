# SA:MP Objects API Service

[https://samp-objects.com](https://samp-objects.com) is a service I built for the San Andreas Multiplayer community - SA:MP 0.3.8 added the ability for servers to provide custom assets to clients so this website is a platform where users can share their work and server owners can search for content to add to their servers.

## Architecture

The site follows a modern, decoupled RESTful service-oriented architecture. The frontend is a React app and [you can check that out here](https://github.com/Southclaws/samp-objects-frontend).

The routes are declared in `router.go` which maps routes to their respective functions. Routes are organised into versions, then namespaces then individual collections - a pretty standard RESTful model. `accounts.go` contains functions for `/v0/accounts/` etc. There are also sub-packages for other isolated components:

- `storage` provides an interface for persistent storage, the current database of choice is MongoDB and the file store is S3-based and uses the Minio client.
- `types` provides common type declarations for structures such as users and objects.

## Authentication

Gorilla is the stack of choice here, it's used for route multiplexing, session state, CSRF tokens and CORS middleware.

Routing is declarative, `Authenticated: true` results in the `app.Authenticated` middleware being applied to a route. This implements JWT authentication (which was implemented early on) and Gorilla Secure Session which was added later on top of JWT and is used to store a user ID with requests to know which user sent each request.

## Development

My personal environment makes heavy use of Docker to provide local services such as databases. If your environment does not have Docker available, you'll have to set your own databases up.

To spin up a local MongoDB and Minio with **very** simple credentials (do not do this on a production server!) simply run:

```bash
make mongodb && make minio
```

MongoDB has no password and Minio uses `default` and `12345678` as the access and secret key.

To fire up a local instance (non containerised) instance of the server, run:

```bash
make local
```

And to fire up a container, run:

```bash
make run
```

For a production deployment, make sure `.env` contains all the production credentials (I am still working on a better way of handling secrets in production!) and run:

```bash
make run-prod
```

Check the `makefile` for other commands.

## Tests

The routing does not have full test coverage yet (it might do in the future, and chances are I'll forget to update this readme when it does!) however the `storage` package does.

## Contributing

I welcome contributions, make sure you follow Go standards (gofmt etc) and run static analysis on your code to ensure quality.
