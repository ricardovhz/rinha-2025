FROM alpine:3.14 AS builder
RUN apk update && apk add ca-certificates && \
  rm -rf /var/cache/apk/*

FROM ubuntu:latest
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY app /opt/app/
EXPOSE 3000
ENTRYPOINT ["/opt/app/app"]

