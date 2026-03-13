#!/usr/bin/env bash

set -e

if [ "$#" -ne 1 ]; then
    echo "usage: $0 user@server-address"
    exit 1
fi

APP_NAME="starlite"
SERVER_SSH=$1
SERVER_PATH=/opt/apps_workspace/$APP_NAME
BINARY_NAME="$APP_NAME-amd64"
SERVER_RESTART_COMMAND="systemctl restart $APP_NAME"
SYSTEMD_FILE="$APP_NAME.service"
SYSTEMD_DAEMONRELOAD_COMMAND="systemctl daemon-reload"

# https://caddyserver.com/docs/running#unit-files
CADDY_RESTART_COMMAND="systemctl reload caddy"
CADDYFILE="$APP_NAME-Caddyfile"

# Assume the script will be run inside the `src/` directory.
OUTFILE="./bin/$BINARY_NAME"
ENVFILENAME=".env.local"
ENVFILE="./bin/$ENVFILENAME"
COMMIT_HASH=$(git rev-parse HEAD)
# COMMIT_HASH="commit_unknown"
BUILD_TIMESTAMP=$(TZ=UTC date -u +"%Y%m%d_%H%M%S")
FILE_HASH=$(b2sum $OUTFILE | cut -f1 -d' ')
REMOTE_FILENAME="$BINARY_NAME-$BUILD_TIMESTAMP-$COMMIT_HASH-$FILE_HASH"

echo "Deploying: $REMOTE_FILENAME"

# Copy necessary files from current version.
scp "$OUTFILE" "$SERVER_SSH:/tmp/$REMOTE_FILENAME"
scp "$ENVFILE" "$SERVER_SSH:/tmp/$REMOTE_FILENAME-$ENVFILENAME"
scp "_tools/files/etc/caddy/sites-enabled/$CADDYFILE" "$SERVER_SSH:/tmp/$REMOTE_FILENAME-$CADDYFILE"
scp "_tools/files/etc/systemd/system/$SYSTEMD_FILE" "$SERVER_SSH:/tmp/$REMOTE_FILENAME-$SYSTEMD_FILE"

# Put the latest files in the right directories and restart everything without downtime.
ssh -q -Tt $SERVER_SSH <<EOL
    sudo nohup sh -c "\
    mkdir -p $SERVER_PATH/versions/ $SERVER_PATH/current/ /etc/caddy/sites-enabled/ && \
    mv "/tmp/$REMOTE_FILENAME-$CADDYFILE" "/etc/caddy/sites-enabled/$CADDYFILE" && \
    $CADDY_RESTART_COMMAND && \
    mv "/tmp/$REMOTE_FILENAME-$SYSTEMD_FILE" "/etc/systemd/system/$SYSTEMD_FILE" && \
    $SYSTEMD_DAEMONRELOAD_COMMAND && \
    mv "/tmp/$REMOTE_FILENAME-$ENVFILENAME" "$SERVER_PATH/current/$ENVFILENAME" && \
    mv "/tmp/$REMOTE_FILENAME" "$SERVER_PATH/versions/$REMOTE_FILENAME" && \
    chmod +x "$SERVER_PATH/versions/$REMOTE_FILENAME" && \
    rm -f "$SERVER_PATH/current/$BINARY_NAME" && \
    ln -s "$SERVER_PATH/versions/$REMOTE_FILENAME" "$SERVER_PATH/current/$BINARY_NAME" && \
    $SERVER_RESTART_COMMAND"
EOL

echo "Deleting older versions, retaining the latest 10!"

# Cleanup old versions, and retain the last 10 deployed.
# In order to retain 10x versions we need to keep the top 10 lines when
# sorted with the latest files at the top, and start removing from line 11!
# Attention: If you have less than 10 deployments already this will fail, but it's fine to ignore.
ssh -q -Tt $SERVER_SSH <<EOL
    sudo nohup sh -c "find "$SERVER_PATH/versions/" -type f -exec realpath {} \; | sort -r | tail -n +11 | sudo xargs rm"
EOL
