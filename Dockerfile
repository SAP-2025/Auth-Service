FROM golang:1.23.2-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o auth-service cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/auth-service .
COPY config.yaml .
EXPOSE 8001
CMD ["./auth-service"]