.PHONY: all test race bench lint catalog catalog-full site cover

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
	go test -coverprofile=/tmp/devicex.cover ./... && go tool cover -func=/tmp/devicex.cover | tail -1

# Rebuild the catalogue from a Google Play Console export. This is the current
# source, updated weekly, and the one to prefer.
#
#   make catalog CSV=supported_devices.csv
catalog:
	@test -n "$(CSV)" || (echo "usage: make catalog CSV=supported_devices.csv"; exit 2)
	go run ./gen -csv "$(CSV)" -source "Google Play supported-devices list (storage.googleapis.com/play_public/supported_devices.csv)" -generated "$$(date +%Y-%m-%d)"
	gofmt -w internal/catalog

# Rebuild from both sources: the Play list, plus the English MobileModels pages
# for the regional variants it omits. This is what ships.
#
# MobileModels is CC BY-NC-SA 4.0 and the generated catalogue inherits it.
#
#   git clone https://github.com/KHwang9883/MobileModels
#   make catalog-full CSV=supported_devices.csv MD=MobileModels/brands
catalog-full:
	@test -n "$(CSV)" -a -n "$(MD)" || (echo "usage: make catalog-full CSV=supported_devices.csv MD=MobileModels/brands"; exit 2)
	go run ./gen -csv "$(CSV)" -md "$(MD)" -web api \
	  -source "Google Play supported-devices list (storage.googleapis.com/play_public/supported_devices.csv), supplemented with KHwang9883/MobileModels (CC BY-NC-SA 4.0)" \
	  -generated "$$(date +%Y-%m-%d)"
	gofmt -w internal/catalog

# Serve the Pages site locally. The API is static files, so this needs nothing
# but a file server; what runs in production is the same directory.
#
#   make site   then open http://localhost:8000/lookup.html#SM-S928B
site:
	python3 -m http.server 8000
