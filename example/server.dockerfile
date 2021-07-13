FROM golang:1.15 as builder
WORKDIR /workdir
COPY ./ ./
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -o server ./cmd/server

FROM alpine:3.14 as certifier
RUN apk add -U --no-cache ca-certificates

FROM scratch
WORKDIR /
COPY --from=builder /workdir/server .
COPY --from=certifier /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/server"]
