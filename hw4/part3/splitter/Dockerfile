# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o splitter main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/splitter .
EXPOSE 8080
CMD ["./splitter"]