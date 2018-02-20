#!/bin/bash

# Start redis if its not running
if  [ "`docker ps -q --filter name=redis-local`" = "" ]; then
   docker run -d -p 6379:6379 --name redis-local redis
fi

# Remove our app and restart it
docker rm -f app-local 2> /dev/null

docker run --rm -it -p 8085:8085 --link redis-local:redis-local --name app-local app-local redis-local
