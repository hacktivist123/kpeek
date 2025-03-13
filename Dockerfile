# Dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o kpeek .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/kpeek /usr/local/bin/kpeek
ENTRYPOINT ["kpeek"]
