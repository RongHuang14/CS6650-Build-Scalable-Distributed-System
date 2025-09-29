# mapper/Dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app

# Copy go mod file
COPY go.mod ./

# Copy source code
COPY main.go .

# Initialize go module and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN go build -o mapper main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/mapper .
EXPOSE 8080
CMD ["./mapper"]