FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o consumer-service ./cmd/consumer

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/consumer-service .
COPY config.json .
CMD ["./consumer-service"]
