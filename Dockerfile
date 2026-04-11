FROM golang:1.22-alpine AS builder
WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/unibee-api main.go

FROM alpine:3.20
WORKDIR /app

RUN apk add --no-cache \
    ca-certificates \
    curl \
    tzdata \
    fontconfig \
    ttf-dejavu

COPY --from=builder /app/bin/unibee-api /app/main
COPY config.yaml /app/config.yaml
COPY version.txt /app/version.txt
COPY resource /app/resource
COPY manifest/i18n /app/i18n
COPY manifest/fonts /usr/share/fonts
COPY manifest/docker/entrypoint.coolify.sh /app/entrypoint.sh

RUN chmod +x /app/main /app/entrypoint.sh

EXPOSE 8088

HEALTHCHECK --interval=15s --timeout=5s --start-period=40s --retries=6 \
  CMD curl -fsS http://127.0.0.1:8088/health || exit 1

ENTRYPOINT ["/app/entrypoint.sh"]
