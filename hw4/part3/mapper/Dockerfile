# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go build -o mapper main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mapper .
EXPOSE 8080
CMD ["./mapper"]