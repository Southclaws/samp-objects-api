VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
MONGO_PASS := $(shell cat MONGO_PASS.private)
AUTH_SECRET := $(shell cat AUTH_SECRET.private)
STORE_ACCESS := $(shell cat STORE_ACCESS.private)
STORE_SECRET := $(shell cat STORE_SECRET.private)
STORE_LOCATION := $(shell cat STORE_LOCATION.private)

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
	STORE_HOST=localhost \
	STORE_PORT=9000 \
	STORE_ACCESS=default \
	STORE_SECRET=12345678 \
	STORE_SECURE=false \
	STORE_BUCKET=samp-objects \
	STORE_LOCATION=AMS3 \
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
		-e DOMAIN=localhost \
		-e MONGO_USER=sampobjects \
		-e MONGO_HOST=localhost \
		-e MONGO_PORT=27017 \
		-e MONGO_NAME=sampobjects \
		-e MONGO_PASS=$(MONGO_PASS) \
		-e AUTH_SECRET=$(AUTH_SECRET) \
		-e STORE_HOST=localhost \
		-e STORE_PORT=9000 \
		-e STORE_ACCESS=default \
		-e STORE_SECRET=12345678 \
		-e STORE_SECURE=false \
		-e STORE_BUCKET=samp-objects \
		-e STORE_LOCATION=AMS3 \
		southclaws/samp-objects:$(VERSION)

run-prod:
	-docker rm samp-objects-api
	docker run \
		--name samp-objects-api \
		--restart on-failure \
		-d \
		-p 80:7791 \
		-e BIND=0.0.0.0:80 \
		-e DOMAIN=localhost \
		-e MONGO_USER=sampobjects \
		-e MONGO_HOST=mongodb \
		-e MONGO_PORT=27017 \
		-e MONGO_NAME=sampobjects \
		-e MONGO_PASS=$(MONGO_PASS) \
		-e AUTH_SECRET=$(AUTH_SECRET) \
		-e STORE_HOST=samp-objects.ams3.digitaloceanspaces.com \
		-e STORE_PORT=443 \
		-e STORE_ACCESS=$(STORE_ACCESS) \
		-e STORE_SECRET=$(STORE_SECRET) \
		-e STORE_SECURE=true \
		-e STORE_BUCKET=samp-objects \
		-e STORE_LOCATION=$(STORE_LOCATION) \
		southclaws/samp-objects:$(VERSION)
	docker network connect samp-objects samp-objects-api

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
	docker run \
		--name mongodb \
		-p 27017:27017 \
		-d \
		mongo

minio:
	-docker stop minio
	-docker rm minio
	docker run \
		--name minio \
		-p 9000:9000 \
		-d \
		-e MINIO_ACCESS_KEY=default \
		-e MINIO_SECRET_KEY=12345678 \
		minio/minio server /data
	docker logs minio
