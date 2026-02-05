#!/usr/bin/bash
set -e;

export CURRENT_DIR=$(pwd);
export UNIT_TEST_DIR=$CURRENT_DIR/tests/unit_test;
export BACKEND_API_TEST_DIR=$CURRENT_DIR/tests/backend_api;

# behavioural test first (they create the database!)
cd $BACKEND_API_TEST_DIR;
go test -v;

echo "INFO: test in \"$BACKEND_API_TEST_DIR\" finished";

# now unit tests should pass because database exists
cd $UNIT_TEST_DIR;
go test -v;

echo "INFO: test in \"$UNIT_TEST_DIR\" finished";

cd $CURRENT_DIR;
