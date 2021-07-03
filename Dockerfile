FROM golang:1.12-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*


WORKDIR /app

ENV GIT_TERMINAL_PROMPT 1

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-mysql-elasticsearch -a -ldflags '-extldflags "-static"'  cmd/go-mysql-elasticsearch/main.go


FROM ubuntu:18.04

RUN apt -y update
RUN apt -y install mysql-client

WORKDIR /app
COPY --from=builder /app .

ENTRYPOINT ["./go-mysql-elasticsearch"]
