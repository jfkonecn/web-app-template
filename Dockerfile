FROM node:24-bookworm-slim AS assets

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY web ./web
RUN npm run build:assets

FROM golang:1.25-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY cmd ./cmd
COPY internal ./internal
COPY --from=assets /app/web ./web

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

FROM debian:bookworm-slim AS runtime

WORKDIR /app

RUN groupadd --system app && useradd --system --gid app --create-home app

COPY --from=builder /out/server /app/server
COPY --from=builder /app/web/templates ./web/templates
COPY --from=builder /app/web/static ./web/static

EXPOSE 8080

USER app

ENTRYPOINT ["/app/server"]
