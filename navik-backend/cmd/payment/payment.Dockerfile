FROM golang:1.23-alpine
WORKDIR /app
COPY cmd/main.go ./main.go
RUN go mod init demo && go mod tidy
RUN go build -o paymentservice main.go
EXPOSE 8087
CMD ["./paymentservice"]
