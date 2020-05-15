FROM golang:1.14.2-alpine3.11 as builder

RUN apk add build-base

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build

FROM alpine:3.11

RUN apk add libstdc++

WORKDIR /app

COPY --from=builder /app/ethdo /app

ENTRYPOINT ["/app/ethdo"]