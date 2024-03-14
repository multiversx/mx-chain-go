pushd ../..

docker build -f docker/seednode/Dockerfile . -t seednode:dev

ocker build -f docker/node/Dockerfile . -t node:dev