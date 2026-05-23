FROM golang:1.22-alpine AS builder

WORKDIR /app


COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o bin/server ./cmd/server

FROM alpine:latest

# Zaman dilimi ve SSL sertifikaları için gerekli
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/bin/server .

EXPOSE 8000
CMD ["./server"]