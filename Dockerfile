FROM golang:alpine as base

FROM scratch

COPY --from=base /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./output ./event-tracking-service

CMD ["/event-tracking-service"]
