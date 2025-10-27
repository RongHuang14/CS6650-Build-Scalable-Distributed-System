FROM --platform=linux/amd64 golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o order-processor main.go

FROM --platform=linux/amd64 alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/order-processor .
CMD ["./order-processor"]