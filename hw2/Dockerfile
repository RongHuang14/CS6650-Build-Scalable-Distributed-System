# Step 1: Build the Go app
FROM golang:1.23 AS build

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum first (if you had them)
# COPY go.mod go.sum ./
# RUN go mod download

# Copy source code
COPY . .

# Build the Go app
RUN go build -o server .

# Step 2: Run the app in a lightweight image
FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]