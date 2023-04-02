#!/usr/bin/env bash

while IFS= read -r -d '' file
do
    protoc -I /usr/local/include -I ./api \
      --go_out=./api --go_opt=paths=source_relative \
      --go-grpc_out=./api --go-grpc_opt=paths=source_relative \
      --go-grpc_opt=require_unimplemented_servers=false \
      "$file"
done <  <(find './api' -type f -name "*.proto" -print0)