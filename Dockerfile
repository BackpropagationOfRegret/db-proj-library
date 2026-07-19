FROM golang:1.26-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/library ./cmd/library \
 && CGO_ENABLED=0 GOOS=linux go build -o /out/seed ./cmd/seed

FROM debian:bookworm-slim

RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates curl \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=builder /out/library /app/library
COPY --from=builder /out/seed /app/seed
COPY migrations /app/migrations

ENV DATABASE_URL=postgres://library:library@postgres:5432/library?sslmode=disable \
    HTTP_ADDR=:8080 \
    MIGRATIONS_PATH=file://migrations

EXPOSE 8080
CMD ["/app/library"]
