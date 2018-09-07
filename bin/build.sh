#!/bin/bash -eu

run() {
	fn="$1"

	echo "# $fn"

	"$fn"
}

downloadDependencies() {
	dep ensure
}

checkFormatting() {
	# unfortunately we need to list formattable directories because "." would include vendor/
	local offenders=$(gofmt -l main.go adapters/ hapitypes/ libraries/ util/)

	if [ ! -z "$offenders" ]; then
		>&2 echo "formatting errors: $offenders"
		exit 1
	fi
}

unitTests() {
	go test ./...
}

staticAnalysis() {
	go vet ./...
}

buildLinuxArm() {
	GOOS=linux GOARCH=arm go build -o rel/home-automation-hub_linux-arm
}

buildLinuxAmd64() {
	GOOS=linux GOARCH=amd64 go build -o rel/home-automation-hub_linux-amd64
}

uploadBuildArtefacts() {
	# the CLI breaks automation unless opt-out..
	export JFROG_CLI_OFFER_CONFIG=false

	jfrog-cli bt upload \
		"--user=joonas" \
		"--key=$BINTRAY_APIKEY" \
		--publish=true \
		'rel/*' \
		"function61/home-automation-hub/main/$FRIENDLY_REV_ID" \
		"$FRIENDLY_REV_ID/"
}

rm -rf rel
mkdir rel

run downloadDependencies

run checkFormatting

run staticAnalysis

run unitTests

run buildLinuxArm

run buildLinuxAmd64

if [ "${PUBLISH_ARTEFACTS:-''}" = "true" ]; then
	run uploadBuildArtefacts
fi

