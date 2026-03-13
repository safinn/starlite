# Starlite

Boilerplate for creating real-time Hypermedia applications in Go. Heavily inspired by [northstar](https://github.com/zangster300/northstar).

## What is included?

- [Datastar](https://data-star.dev/)
- SQLite3 with [sqlc](https://sqlc.dev/) and [sqlc-gen-zombiezen](https://github.com/delaneyj/toolbelt/tree/main/sqlc-gen-zombiezen) for database storage
- [templ](https://templ.guide/) for powerful HTML templates
- [Embedded nats](https://github.com/delaneyj/toolbelt/tree/main/embeddednats) as an event bus
- [Mise](https://mise.jdx.dev/) for tool management and task runner
- Hot reloading in development using [Air](https://github.com/air-verse/air) and [watchexec](https://github.com/watchexec/watchexec)
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

### Deploy

`_tools/deploy.sh` performs a versioned SSH rollout. It uploads:

- `bin/<binary-name>`
- `bin/.env.local`
- `_tools/files/etc/systemd/system/<service-name>.service`
- `_tools/files/etc/caddy/sites-enabled/<site-name>-Caddyfile`

Then it:

- Places release binaries in `/opt/apps_workspace/<app-name>/versions/` (filename includes build timestamp, commit hash, and file hash)
- Updates `/opt/apps_workspace/<app-name>/current/<binary-name>` symlink
- Reloads Caddy and systemd, then restarts the app service
- Prunes old releases (keeps latest 10)

Prerequisites:

1. Run one-time server provisioning (on the server as root):
```sh
sudo bash _tools/setup.sh <app-name>
```
   Run this from an existing root SSH session. The setup script bootstraps the entire server:
   - Updates system packages and installs essentials (`ufw`, `fail2ban`, `sudo`, `unattended-upgrades`, etc.)
   - Configures UFW firewall (allows ports 22, 80, 443)
   - Creates a `deployer` user (with passwordless sudo) and an `<app-name>` service user
   - Copies root's `authorized_keys` to the deployer user
   - Hardens SSH (disables password auth, restricts login to `deployer` only)
   - Configures fail2ban for SSH brute-force protection
   - Installs Caddy from the official repository
   - Creates app directories under `/opt/apps_workspace/<app-name>/`
2. Update `APP_NAME` in `_tools/deploy.sh` for your app (all other constants are derived from it).
3. Ensure the deploy templates match your server:
   - `_tools/files/etc/systemd/system/<service-name>.service` (user, paths, service name)
   - `_tools/files/etc/caddy/sites-enabled/<site-name>-Caddyfile` (domain and upstream port)
4. Ensure `bin/.env.local` exists before deploy.

Deploy:

```sh
bash _tools/deploy.sh deployer@<server-ip>
```

### Docker

Build the image:

```sh
docker build -t <app-name>:latest .
```

Run the app (default `ENV` and `PORT` are set in `Dockerfile`):

```sh
docker run --rm --name <app-name> -p 9001:9001 <app-name>:latest
```

Persist runtime data (SQLite/NATS files) with a volume:

```sh
docker run --rm --name <app-name> -p 9001:9001 -v <app-name>-data:/data <app-name>:latest
```
