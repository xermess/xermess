# The xermess API. Build context: the repository root.
#
#   docker build -f deploy/docker/api.Dockerfile -t xermess-api .

FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd ./cmd
COPY internal ./internal
COPY migrations ./migrations
# The shipped translations are a Go package (locales/locales.go) that embeds
# the JSON beside it, and internal/api/respond imports it.
COPY locales ./locales
ARG VERSION=dev
RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/xermess ./cmd/xermess

FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H -u 10001 xermess
WORKDIR /app
COPY --from=build /out/xermess ./xermess
# The migrations are compiled in and applied on start; goose only needs the
# directory it is pointed at to exist.
RUN mkdir migrations
USER xermess
ENV GIN_MODE=release XERMESS_ADDR=:8080 XERMESS_ADMIN_ADDR=:8081
# 8080 public, 8081 admin: neither is published; Caddy reaches them.
EXPOSE 8080 8081
ENTRYPOINT ["/app/xermess"]
