#!/bin/bash -eu

source /build-common.sh

BINARY_NAME="home-automation-hub"
BINTRAY_PROJECT="function61/home-automation-hub"
GOFMT_TARGETS="main.go adapters/ hapitypes/ libraries/"

standardBuildProcess
