#!/usr/bin/sh
set -e;

mkdir -p public;
mkdir -p containers/postgresql/data;

export TARGET_DIR="$(pwd)/bin";

echo "NOTE: all build goes to $TARGET_DIR";

export BACKEND_API_SOURCE="$(pwd)/cmd/backend_api";
export BACKEND_API_TARGET="$TARGET_DIR/backend_api/main";
export PRODUCER_CTL_SOURCE="$(pwd)/cmd/producer_ctl";
export PRODUCER_CTL_TARGET="$TARGET_DIR/producer_ctl/main";

echo "building: $BACKEND_API_SOURCE";
echo "- target: $BACKEND_API_TARGET";
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o $BACKEND_API_TARGET $BACKEND_API_SOURCE;

echo "building: $PRODUCER_CTL_SOURCE";
echo "- target: $PRODUCER_CTL_TARGET";
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -o $PRODUCER_CTL_TARGET $PRODUCER_CTL_SOURCE;

