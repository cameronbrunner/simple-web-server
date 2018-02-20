#!/bin/bash
set -x

# Just run docker build
docker build . -t app-local
