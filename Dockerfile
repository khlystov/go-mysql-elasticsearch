FROM golang:1.12-alpine AS builder

RUN apk update && apk add alpine-sdk git && rm -rf /var/cache/apk/*


WORKDIR /app

ENV GIT_TERMINAL_PROMPT 1

COPY . .
RUN go build -o go-mysql-elasticsearch cmd/go-mysql-elasticsearch/main.go


FROM alpine:latest

WORKDIR /app
COPY --from=builder /app .

ENTRYPOINT ["./go-mysql-elasticsearch"]
