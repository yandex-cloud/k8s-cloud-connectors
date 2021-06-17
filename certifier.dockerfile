FROM golang:1.15 as builder
WORKDIR /workdir
COPY ./ ./
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o certifier ./cmd/yc-connector-certifier/main.go

FROM alpine:3.14
WORKDIR /
RUN apk add --update openssl && \
        rm -rf /var/cache/apk/*
COPY --from=builder /workdir/certifier .
ENTRYPOINT ["/certifier"]