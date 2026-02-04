#!/usr/bin/bash
set -e;

export CURRENT_DIR=$(pwd);
export UNIT_TEST_DIR=$CURRENT_DIR/tests/unit_test;
export BACKEND_API_TEST_DIR=$CURRENT_DIR/tests/backend_api;

# assertion & data test
cd $UNIT_TEST_DIR;
go test -v;

echo "INFO: test in \"$UNIT_TEST_DIR\" finished";

# behavioural test
cd $BACKEND_API_TEST_DIR;
go test -v;

echo "INFO: test in \"$BACKEND_API_TEST_DIR\" finished";
# echo "INFO: test in \"$BACKEND_API_TEST_DIR\" need to run manually after backend running";

cd $CURRENT_DIR;

