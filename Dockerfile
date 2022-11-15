
# build image
FROM golang:1.18.8-alpine3.16 as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/cortex-gateway cmd/gateway/main.go
RUN CGO_ENABLED=0 go build -o /go/bin/cortex-gateway-tool cmd/tool/main.go

# executable image
FROM alpine:3.16.3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/cortex-gateway /go/bin/cortex-gateway
COPY --from=builder /go/bin/cortex-gateway-tool /go/bin/cortex-gateway-tool

ENV VERSION 0.1.0
ENTRYPOINT ["/go/bin/cortex-gateway"]
