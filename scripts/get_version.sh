#!/bin/bash
SELF_DIR=$(SELF=$(dirname "$0") && bash -c "cd \"$SELF\" && pwd")
GIT_VERSION=$(git describe --tags --abbrev=0 || echo "v0.0.0")
echo -n "$GIT_VERSION" > "$SELF_DIR/version.txt"
