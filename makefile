VERSION := $(shell cat VERSION)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
-include .env
.PHONY: version


# -
# Local
# -


fast:
	go build $(LDFLAGS) -o samp-objects-api

static:
	CGO_ENABLED=0 GOOS=linux go build -a $(LDFLAGS) -o samp-objects-api .

local: fast
	./samp-objects-api

version:
	git tag $(VERSION)
	git push
	git push origin $(VERSION)

test:
	go test -v -race


# -
# Docker
# -


build:
	docker build --no-cache -t southclaws/samp-objects-api:$(VERSION) .

push: build
	docker push southclaws/samp-objects-api:$(VERSION)
	
run:
	-docker rm samp-objects-test
	docker run \
		--name samp-objects-test \
		--network host \
		--env-file .env \
		southclaws/samp-objects-api:$(VERSION)

run-prod:
	-docker stop samp-objects-api
	-docker rm samp-objects-api
	docker run \
		--name samp-objects-api \
		--restart on-failure \
		-d \
		-p 7791:80 \
		--env-file .env \
		southclaws/samp-objects-api:$(VERSION)
	docker network connect samp-objects samp-objects-api

enter:
	docker run -it --entrypoint=bash southclaws/samp-objects-api:$(VERSION)

enter-mount:
	docker run -v $(shell pwd)/testspace:/samp -it --entrypoint=bash southclaws/samp-objects-api:$(VERSION)


# -
# Testing
# -


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
