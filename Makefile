# MAINTAINER: David López Becerra <not4rent@gmail.com>

APP     = filterer
VERSION = v0.1.0

.PHONY: build
build: api-build app-build docker-build

.PHONY: api-build
api-build:
	find . -type f -name '*.pb.go' -delete
	docker compose up -d --build api-builder
	docker compose exec api-builder ./api/generate.sh
	docker compose cp api-builder:/go/src/github.com/lopezator/filterer/api .

.PHONY: app-build
app-build:
	docker compose exec -e GOOS=linux -e GOARCH=amd64 api-builder go build -o /usr/local/bin/filterer ./cmd/$(APP)
	docker compose cp api-builder:/usr/local/bin/filterer ./bin/$(APP)-linux-amd64

.PHONY: docker-build
docker-build:
	cp bin/$(APP)-linux-amd64 build/container/$(APP)-linux-amd64
	chmod 0755 build/container/$(APP)-linux-amd64
	docker build -f build/container/Dockerfile -t filterer/$(APP):$(VERSION) build/container/

.PHONY: release
release: build
	docker push filterer/$(APP):$(VERSION)

.PHONY: clean
clean:
	find . -type f -name 'filterer-linux-amd64' -delete
	docker compose down

.PHONY: test
test: build
	docker compose exec api-builder go test ./...

.PHONY: serve
serve: build
	docker compose up -d filterer