VERSION := $(shell git describe --tags)
BUILD := $(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X=main.Version=$(VERSION) -X=main.Build=$(BUILD)"
MAKEFLAGS += --silent

test:
	go test -race -cover ./...

deps:
	docker buildx create --name builder --driver docker-container --use

clean:
	rm -rf bin/
	go clean
	docker buildx stop builder
	docker buildx rm builder

build:
	go build $(LDFLAGS) -o bin/ ./...

run:
	GODOC_ROOT=/tmp LOG_LEVEL=DEBUG go run main.go

publish:
	docker buildx build --push --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --build-arg VERSION=$(VERSION) --build-arg=BUILD=$(BUILD) --tag ctxsh/gdoc:$(VERSION) --tag ctxsh/gdoc:latest .

docker:
	docker build --build-arg VERSION=$(VERSION) --build-arg=BUILD=$(BUILD) --tag ctxsh/gdoc .
