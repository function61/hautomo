#!/bin/bash -eu

source /build-common.sh

COMPILE_IN_DIRECTORY="cmd/hautomo"
BINARY_NAME="hautomo"
GOFMT_TARGETS="cmd/ pkg/"

standardBuildProcess
