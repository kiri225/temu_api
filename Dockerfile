# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS builder

WORKDIR /src

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/playground ./cmd/playground

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget \
    && adduser -D -H -u 65532 app

WORKDIR /app

COPY --from=builder /out/playground ./playground
COPY cmd/playground/unavailable.json ./data/unavailable.json

RUN chown -R app:app /app

USER app

EXPOSE 8080

ENTRYPOINT ["./playground"]
CMD ["-port", "8080", "-config", "/config/config.json", "-unavailable", "/data/unavailable.json"]
