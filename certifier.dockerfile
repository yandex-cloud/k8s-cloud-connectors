FROM golang:1.15 as builder
WORKDIR /workdir
COPY ./ ./
RUN go mod download && apt-get install openssl && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o certifier ./cmd/yc-connector-certifier/main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workdir/certifier .
ENTRYPOINT ["/certifier"]