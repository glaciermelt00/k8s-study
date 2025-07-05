FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o migration-cli cmd/migrate/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/migration-cli .
COPY --from=builder /app/migrations ./migrations

CMD ["./migration-cli"]
