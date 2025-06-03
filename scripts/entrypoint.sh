#!/bin/sh

# Ensure APP_ENV is set
if [ -z "$APP_ENV" ]; then
  APP_ENV=dev
fi

# Construct the variable name dynamically
ENV_PREFIX=$(echo "$APP_ENV" | tr '[:lower:]' '[:upper:]')
VERTEX_AI_CONFIG_VAR="${ENV_PREFIX}_ENV_VERTEX_AI_CONFIG_PATH"

# Get the value of the dynamically constructed variable name
VERTEX_AI_CONFIG_PATH=$(eval echo \$$VERTEX_AI_CONFIG_VAR)

# Check if the constructed environment variable is set
if [ -z "$VERTEX_AI_CONFIG_PATH" ]; then
  echo "Error: ${VERTEX_AI_CONFIG_VAR} is not set."
  exit 1
fi

# Create the directory if it does not exist
DIR=$(dirname "$VERTEX_AI_CONFIG_PATH")
if [ ! -d "$DIR" ]; then
  mkdir -p "$DIR"
fi

# Decode the base64 encoded service account and write to the file
echo $VERTEX_AI_SERVICE_ACCOUNT_BASE64 | base64 -d > "$VERTEX_AI_CONFIG_PATH"

# Proceed with the rest of the entrypoint script or the application start
exec "$@"