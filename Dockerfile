FROM golang:latest as builder

ENV CGO_ENABLED=0

# Install script isn't cross-platform
RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR /go/src/github.com/parrotmac/water-proxy/
COPY Gopkg.toml .
COPY Gopkg.lock .

RUN dep ensure -v -vendor-only
COPY . .
RUN go build -o bin/proxy cmd/main.go
# Final artifact: /go/src/github.com/parrotmac/water-proxy/bin/proxy

FROM alpine:latest

EXPOSE 8000

# Ensure runtime has valid certs
RUN apk update \
        && apk upgrade \
        && apk add --no-cache \
        ca-certificates \
        && update-ca-certificates 2>/dev/null || true

WORKDIR /opt/water-proxy
COPY --from=builder /go/src/github.com/parrotmac/water-proxy/bin/proxy .

CMD ["/opt/water-proxy/proxy"]
