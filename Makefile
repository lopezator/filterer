# MAINTAINER: David López Becerra <not4rent@gmail.com>

APP     = filterer
VERSION = v0.1.0

.PHONY: build
build: api-build

.PHONY: api-build
api-build:
	find . -type f -name '*.pb.go' -delete
	docker compose up -d --build filterer
	docker compose exec filterer ./api/generate.sh
	docker compose cp filterer:/go/src/github.com/lopezator/filterer/api .