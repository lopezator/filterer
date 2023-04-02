FROM golang:1.21.4-bookworm

# Set my module as private.
ENV GOPRIVATE github.com/lopezator/filterer

# Install dependencies.
RUN apt-get update && \
  apt-get install -y --no-install-recommends unzip && \
  rm -rf /var/lib/apt/lists/*

# Install go proto dependencies.
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.3.0

# Install protocol buffers.
ENV PROTOC_VERSION=25.1
ENV PROTOC_ZIP=protoc-$PROTOC_VERSION-linux-x86_64.zip
RUN curl -OL https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_ZIP
RUN unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
RUN unzip -o $PROTOC_ZIP -d /usr/local 'include/*'
RUN rm -f $PROTOC_ZIP

# Install googleapis.
ENV GOOGLEAPIS_SHA=4eccaaf48c0ccadc6f98707d3dbe9614d47bb103
ENV GOOGLEAPIS_ZIP=$GOOGLEAPIS_SHA.zip
RUN curl -OL https://github.com/googleapis/googleapis/archive/$GOOGLEAPIS_ZIP
RUN unzip -oj $GOOGLEAPIS_ZIP -d /usr/local/include/google/api 'googleapis-'$GOOGLEAPIS_SHA'/google/api/*'
RUN rm -f $GOOGLEAPIS_ZIP

# Copy current workspace.
ENV PKGPATH github.com/lopezator/filterer
WORKDIR ${GOPATH}/src/${PKGPATH}
COPY . ${GOPATH}/src/${PKGPATH}