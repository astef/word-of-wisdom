FROM golang:1.21.5 AS builder

WORKDIR /build

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd/server

FROM scratch

COPY --from=builder /build/server /

EXPOSE 5000

ENTRYPOINT ["/server"]
