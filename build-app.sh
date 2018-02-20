#!/bin/bash

# Get our GOOS
export GOOS=$( uname -s | tr '[:upper:]' '[:lower:]' )

# Build our app for our host environment
docker run --rm -it -v `pwd`:/go/src/app -e GOOS=$GOOS cameronbrunner/townhall-builder-template
mv app app-$GOOS

# Always build linux too
if [ "$GOOS" != "linux" ]; then
   docker run --rm -it -v `pwd`:/go/src/app -e GOOS=linux cameronbrunner/townhall-builder-template
   mv app app-linux
fi
