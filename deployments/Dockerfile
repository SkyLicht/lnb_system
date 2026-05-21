FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal

RUN go test ./... && CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /bin/file-watcher ./cmd/file-watcher

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=build /bin/file-watcher /usr/local/bin/file-watcher
COPY configs ./configs

RUN mkdir -p /app/logs /app/watched/input /app/watched/output && chown -R app:app /app

USER app

ENTRYPOINT ["file-watcher"]
CMD ["-config", "configs/watcher.config.json"]
