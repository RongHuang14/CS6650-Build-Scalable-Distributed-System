# splitter/Dockerfile
# Build stage
FROM --platform=$BUILDPLATFORM golang:1.23-alpine AS builder
WORKDIR /app

# Enable static build by default
ENV CGO_ENABLED=0

# Target platform (provided by buildx)
ARG TARGETOS
ARG TARGETARCH

# Copy go mod file
COPY go.mod ./

# Copy source code
COPY main.go .

# Initialize go module and download dependencies
RUN go mod tidy && go mod download

# Build the application for the target platform
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o splitter main.go

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/splitter .
EXPOSE 8080
CMD ["./splitter"]