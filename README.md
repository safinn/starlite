# Starlite

Boilerplate for creating real-time Hypermedia applications in Go. Heavily inspired by [northstar](https://github.com/zangster300/northstar).

## What is included?

- [Datastar](https://data-star.dev/)
- SQLite3 with [sqlc](https://sqlc.dev/) and [sqlc-gen-zombiezen](https://github.com/delaneyj/toolbelt/tree/main/sqlc-gen-zombiezen) for database storage
- [templ](https://templ.guide/) for powerful HTML templates
- [Embedded nats](https://github.com/delaneyj/toolbelt/tree/main/embeddednats) as an event bus
- [Mise](https://mise.jdx.dev/) for tool management and task runner
- Hot reloading in development using [watchexec](https://github.com/watchexec/watchexec)
- [hashfs](https://github.com/benbjohnson/hashfs) for aggressive HTTP caching of static files

## Getting started

1. Clone the repository
2. Install [`mise`](https://mise.jdx.dev/getting-started.html)
3. Install tools (may need to trust the config first using `mise trust`)
```sh
mise i
```
4. Run in development
```sh
mise dev
```
5. Navigate to [http://localhost:8080](http://localhost:8080) in your favorite web browser

## Debugging

- A Visual Studio Code `Debug Main` configuration in [launch.json](./.vscode/launch.json) is included to launch the application with debugging enabled through the IDE's debugging interface.
- `mise debug` can be used to launch the application with delve (dlv) connected and debugging managed through dlv directly in the terminal.

## Deployment

### Build

```sh
mise build
```

The build command will output binaries in the `./bin` directory.

#### Docker

ko.build is used to build the docker image without requiring docker.

```sh
mise build:image --local # to load into the local docker daemon
KO_DOCKER_REPO=ghcr.io/my-org/my-repo mise build:image # to build and push the image to the container image registry
```

See https://ko.build/get-started/ for authenticating to the container image registry.
