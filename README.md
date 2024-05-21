# filterer

## Description

This is not yet usable, just a PoC for now.

Serving, because of using connect protocol, supports gRPC and REST out of the box.

## Call examples

### Connect (using buf curl)

```bash
buf curl --schema . --protocol connect \
--data '{"expr": "display_name == '\'paco\''"}' \
http://localhost:1337/lopezator.filterer.v1.FiltererService/Filter --http2-prior-knowledge
```

> Note: Example is using `connect` protocol (default if protocol ommited), but `grpc` or `grpcweb` protocols are also possible.

> Note2: If you want to grab the schema from the BSR use `--schema --schema buf.build/lopezator/filterer` instead. 

### REST call

```bash
curl --header "Content-Type: application/json" \
--data '{"expr": "display_name == '\'paco\''"}' \ 
http://localhost:1337/lopezator.filterer.v1.FiltererService/Filter
```

### gRPC (using grpcurl)

```bash
grpcurl \
-protoset <(buf build -o -) \
-plaintext \
-format json \
-d '{"expr": "display_name == '\''paco'\''"}' \
localhost:1337 lopezator.filterer.v1.FiltererService/Filter
```

## Response 

```json
{
  "where": "WHERE: LOWER(display_name) = (LOWER(?)), ARGS: [paco]"
}
```