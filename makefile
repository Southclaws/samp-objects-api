VERSION := $(shell git rev-parse HEAD)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
MONGO_PASS := $(shell cat MONGO_PASS.private)
AUTH_SECRET := $(shell cat AUTH_SECRET.private)

.PHONY: version

fast:
	go build $(LDFLAGS) -o samp-objects-api

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o samp-objects-api .

local: fast
	BIND=localhost:8080 \
	DOMAIN=localhost \
	MONGO_USER=sampobjects \
	MONGO_HOST=localhost \
	MONGO_PORT=27017 \
	MONGO_NAME=sampobjects \
	AUTH_SECRET=$(AUTH_SECRET) \
	DEBUG=1 \
	./samp-objects-api

version:
	git tag $(VERSION)
	git push
	git push origin $(VERSION)

test:
	go test -v -race

# Docker

build:
	docker build --no-cache -t southclaws/samp-objects:$(VERSION) -f Dockerfile.dev .

build-prod:
	docker build --no-cache -t southclaws/samp-objects:$(VERSION) .

build-test:
	docker build --no-cache -t southclaws/samp-objects-test:$(VERSION) -f Dockerfile.testing .

push: build-prod
	docker push southclaws/samp-objects:$(VERSION)
	
run:
	-docker rm samp-objects-test
	docker run \
		--name samp-objects-test \
		--network host \
		-e BIND=localhost:8080 \
		-e MONGO_USER=sampobjects \
		-e MONGO_HOST=localhost \
		-e MONGO_PORT=27017 \
		-e MONGO_NAME=sampobjects \
		-e AUTH_SECRET=$(AUTH_SECRET) \
		-
		southclaws/samp-objects:$(VERSION)

enter:
	docker run -it --entrypoint=bash southclaws/samp-objects:$(VERSION)

enter-mount:
	docker run -v $(shell pwd)/testspace:/samp -it --entrypoint=bash southclaws/samp-objects:$(VERSION)

# Test stuff

test-container: build-test
	docker run --network host southclaws/samp-objects-test:$(VERSION)

mongodb:
	-docker stop mongodb
	-docker rm mongodb
	docker run --name mongodb -p 27017:27017 -d mongo
