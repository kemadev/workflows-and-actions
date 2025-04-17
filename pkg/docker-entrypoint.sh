#!/usr/bin/env bash

main() {
	cd /tool/pkg/pkg/ci/run || {
		echo "Failed to change directory to /tool/pkg/pkg/ci/run"
		exit 1
	}
	go run .
}

main "${@}"
