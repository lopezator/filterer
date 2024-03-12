FROM golang:1.21.4-bookworm AS build

# Install dependencies.
RUN apt-get update && \
  apt-get install -y --no-install-recommends && \
  rm -rf /var/lib/apt/lists/*

# Copy current workspace.
WORKDIR /go/src/github.com/lopezator/filterer
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o /go/bin/filterer ./cmd/filterer

# Now copy it into our base image.
FROM gcr.io/distroless/static-debian12 AS prod
COPY --from=build /go/bin/filterer /
ENTRYPOINT ["/filterer"]