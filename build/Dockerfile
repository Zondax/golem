# syntax=docker/dockerfile:1
FROM golang:1.24.1-alpine AS builder

RUN apk update && apk --no-cache --update add build-base bash git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.17 AS production

RUN apk update && apk --no-cache --update add ca-certificates

RUN addgroup --system --gid 1001 zondax
RUN adduser --system --uid 1001 zondax
USER zondax

COPY --chown=zondax:zondax --from=builder /app/output /zondax/bin

EXPOSE 9090

CMD ["/zondax/bin/golem", "start", "-c", "/zondax/config/golem.yaml"]
