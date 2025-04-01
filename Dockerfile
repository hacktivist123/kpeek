# Dockerfile
FROM alpine:3.18
COPY kpeek /usr/local/bin/kpeek
ENTRYPOINT ["kpeek"]
