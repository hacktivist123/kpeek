# Dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags "-X github.com/hacktivist123/kpeek/cmd.version=${KPEEK_VERSION}" -o kpeek .

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/kpeek /usr/local/bin/kpeek
ENTRYPOINT ["kpeek"]
