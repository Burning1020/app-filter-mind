.PHONY: build test clean docker

GO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=cmd/app-filter-mind
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X app-filter-mind.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

cmd/app-filter-mind:
	$(GO) build $(GOFLAGS) -o $@ ./cmd

run:
	cd bin && ./edgex-launch.sh

docker:
	docker build \
		--label "git_sha=$(GIT_SHA)" \
		-t burning1020/docker-app-filter-mind:$(VERSION)-dev \
		.

test:
	$(GO) test ./... -cover
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(MICROSERVICES)
