FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 go build -o producer-service ./cmd/producer

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/producer-service .
COPY config.json .
EXPOSE 8080
CMD ["./producer-service"]
