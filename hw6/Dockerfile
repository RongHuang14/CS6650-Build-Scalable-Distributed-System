# hw6/Dockerfile
FROM --platform=linux/amd64 python:3.11-slim

WORKDIR /app

# Install curl for health checks
RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

# Copy and install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy source code
COPY src/ ./src/

EXPOSE 8080

# Run the application
CMD ["python", "src/main.py"]