#!/usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
CMD="bash"
CMD_ARGS="scripts/scripts_in_docker/build-protos.sh"

TAG="protoc"

docker build -t $TAG -f $DIR/Dockerfile.protoc $DIR
docker run \
       --rm -v $PWD:/src:rw,Z -u $(id -u):$(id -g) --workdir /src \
       --entrypoint $CMD $TAG $CMD_ARGS
