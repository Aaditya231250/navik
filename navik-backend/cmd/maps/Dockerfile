FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/search-service ./cmd/main.go

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/search-service .

EXPOSE 8080

CMD ["/app/search-service"]
