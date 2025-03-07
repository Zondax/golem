#!/bin/bash
SELF_DIR=$(SELF=$(dirname "$0") && bash -c "cd \"$SELF\" && pwd")
GIT_REVISION_1=$(git rev-parse --abbrev-ref HEAD)
GIT_REVISION_2=$(git rev-parse --short HEAD)
echo -n "$GIT_REVISION_1-$GIT_REVISION_2" > "$SELF_DIR/revision.txt"
