FROM golang:1.15 as builder
WORKDIR /workdir
COPY ./ ./
RUN go mod download && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o server ./cmd/server/main.go

FROM scratch
WORKDIR /
COPY --from=builder /workdir/server .
ENTRYPOINT ["/server"]
