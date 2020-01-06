#!/usr/bin/env bash
set -ex

# Adhering to ginkgo's default threshold
SLOW_SPEC_THRESHOLD=5

go get -u golang.org/x/lint/golint
go get -u github.com/onsi/ginkgo/ginkgo

ginkgo \
    -r \
    -race \
    -randomizeAllSpecs \
    -randomizeSuites \
    -keepGoing \
    -slowSpecThreshold="${SLOW_SPEC_THRESHOLD}"

golint ./...