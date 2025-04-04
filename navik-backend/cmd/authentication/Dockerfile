
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc g++ make musl-dev git ca-certificates tzdata

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o auth-service ./cmd/main.go

FROM alpine:3.18

RUN apk update && \
    apk add --no-cache ca-certificates tzdata curl jq && \
    update-ca-certificates

WORKDIR /app

COPY --from=builder /app/auth-service .
# COPY --from=builder /app/config.json ./config.json

# RUN addgroup -S appgroup && adduser -S appuser -G appgroup
# RUN chown -R appuser:appgroup /app

# ENV SERVER_PORT="8080" \
#     DYNAMODB_TABLE_NAME="Users" \
#     DYNAMODB_ENDPOINT="http://dynamodb-local:8000" \
#     AWS_REGION="us-east-1" \
#     JWT_ACCESS_SECRET="" \
#     JWT_REFRESH_SECRET="" \
#     JWT_ACCESS_EXPIRY_MINUTES="15" \
#     JWT_REFRESH_EXPIRY_DAYS="7" \
#     JWT_ISSUER="auth-service"

# USER appuser
EXPOSE 8080


CMD ["./auth-service"]