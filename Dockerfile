# syntax=docker/dockerfile:1

FROM golang:1.24.7-alpine AS builder

ARG APP_VERSION="undefined"
ARG BUILD_TIME="undefined"

WORKDIR /go/src/github.com/artarts36/postgres-backup

RUN apk add git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X 'main.Version=${APP_VERSION}' -X 'main.BuildDate=${BUILD_TIME}'" -o /go/bin/postgres-backup /go/src/github.com/artarts36/postgres-backup/cmd/postgres-backup/main.go

######################################################

FROM postgres:18-alpine

RUN rm -rf /var/lib/postgresql/data

COPY --from=builder /go/bin/postgres-backup /go/bin/postgres-backup

ENTRYPOINT ["/go/bin/postgres-backup"]
