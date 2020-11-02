#!/bin/bash
set -e

mkdir -p /src/app/frontend/.cache/yarn
export YARN_CACHE_FOLDER=/src/app/frontend/.cache/yarn

cd app/frontend && yarn && yarn build && cd -
cp -rT app/frontend/dist build/frontend
