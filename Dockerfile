FROM golang:1.14-buster as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build

FROM debian:buster-slim

RUN mkdir /data
WORKDIR /app

COPY --from=builder /app/ethdo /app

ENTRYPOINT ["/app/ethdo", "--basedir=/data"]
