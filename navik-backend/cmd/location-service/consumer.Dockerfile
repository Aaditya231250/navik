FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc g++ make musl-dev pkgconfig librdkafka-dev 

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=1 go build -tags musl -o consumer-service ./cmd/consumer

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/consumer-service .
COPY config.json .
COPY scripts/ /scripts/

RUN apk add --no-cache bash curl kafkacat librdkafka

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN chown -R appuser:appgroup /app /scripts

USER appuser

CMD ["./consumer-service"]
