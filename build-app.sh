#!/bin/bash

# Toolchain TAG
TAG="${TAG:-latest}"

# Get our GOOS
export GOOS=$( uname -s | tr '[:upper:]' '[:lower:]' )

# Build our app for our host environment
set -x
docker run --rm -it -v `pwd`:/go/src/app -e GOOS=$GOOS cameronbrunner/townhall-builder-template:$TAG
set +x
mv app app-$GOOS

# Always build linux too
if [ "$GOOS" != "linux" ]; then
   set -x
   docker run --rm -it -v `pwd`:/go/src/app -e GOOS=linux cameronbrunner/townhall-builder-template:$TAG
   set +x
   mv app app-linux
fi
