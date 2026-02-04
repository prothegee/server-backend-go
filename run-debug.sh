#!/usr/bin/bash
set -e;

export CURRENT_DIR=$(pwd);
export BACKEND_API_DIR=$CURRENT_DIR/cmd/backend_api;

cd $BACKEND_API_DIR; dlv debug .;

# finally
cd $CURRENT_DIR;

