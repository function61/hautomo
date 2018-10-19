#!/bin/bash -eu

source /build-common.sh

COMPILE_IN_DIRECTORY="cmd/home-automation-hub"
BINARY_NAME="home-automation-hub"
BINTRAY_PROJECT="function61/home-automation-hub"
GOFMT_TARGETS="cmd/ pkg/"

standardBuildProcess
