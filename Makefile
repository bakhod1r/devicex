.PHONY: all test race bench lint catalog catalog-json cover

all: lint test

test:
	go test ./...

race:
	go test -race ./...

bench:
	go test -run=XXX -bench=. -benchmem ./

lint:
	gofmt -l .
	go vet ./...

cover:
	go test -coverprofile=/tmp/adx.cover ./... && go tool cover -func=/tmp/adx.cover | tail -1

# Rebuild the catalogue from a Google Play Console export. This is the current
# source, updated weekly, and the one to prefer.
#
#   make catalog CSV=supported_devices.csv
catalog:
	@test -n "$(CSV)" || (echo "usage: make catalog CSV=supported_devices.csv"; exit 2)
	go run ./gen -csv "$(CSV)" -source "Google Play Console device catalogue export, $$(date +%Y-%m-%d)"
	gofmt -w internal/catalog

# Rebuild from the community JSON list. Older, but needs no Play Console access.
#
#   make catalog-json JSON=devices.json
catalog-json:
	@test -n "$(JSON)" || (echo "usage: make catalog-json JSON=devices.json"; exit 2)
	go run ./gen -json "$(JSON)" -source "pbakondy/android-device-list (MIT), derived from the Google Play supported-devices catalogue"
	gofmt -w internal/catalog
