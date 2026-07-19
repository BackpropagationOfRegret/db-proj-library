FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/library ./cmd/library \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./cmd/seed \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/sync-search ./cmd/sync-search

FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/library /app/library
COPY --from=builder /out/seed /app/seed
COPY --from=builder /out/sync-search /app/sync-search
COPY migrations /app/migrations

ENV DATABASE_URL=postgres://library:library@postgres:5432/library?sslmode=disable \
    HTTP_ADDR=:8080 \
    MIGRATIONS_PATH=file://migrations \
    ELASTICSEARCH_URL=http://elasticsearch:9200 \
    ELASTICSEARCH_INDEX=books

EXPOSE 8080
CMD ["/app/library"]
