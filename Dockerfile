FROM golang:1.17-bullseye as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build

FROM debian:bullseye-slim

RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt install -y ca-certificates && apt-get clean && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/ethdo /app

ENTRYPOINT ["/app/ethdo"]
