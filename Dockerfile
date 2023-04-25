
# build image
FROM golang:1.20.3-alpine3.16 as builder
RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go build -o /bin/cortex-gateway cmd/gateway/main.go
RUN CGO_ENABLED=0 go build -o /bin/cortex-gateway-tool cmd/tool/main.go

# executable image
FROM alpine:3.16.3
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/cortex-gateway /bin/cortex-gateway
COPY --from=builder /bin/cortex-gateway-tool /bin/cortex-gateway-tool
COPY cortex-gateway.yaml /etc/cortex-gateway.yaml

ENV VERSION 0.1.0
ENTRYPOINT ["/bin/cortex-gateway"]
CMD ["-config.file=/etc/cortex-gateway.yaml"]

