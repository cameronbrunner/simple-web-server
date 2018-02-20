#!/bin/bash

# Start redis if its not running
if  [ "`docker ps -q --filter name=redis-local`" = "" ]; then
   docker run -d -p 6379:6379 --name redis-local redis
fi

# Get our GOOS
export GOOS=$( uname -s | tr '[:upper:]' '[:lower:]' )

./app-$GOOS
