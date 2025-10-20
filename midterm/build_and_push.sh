#!/bin/bash

# Build and push Docker images to ECR
# Usage: ./build_and_push.sh <aws_account_id> <aws_region>

set -e

if [ $# -ne 2 ]; then
    echo "Usage: $0 <aws_account_id> <aws_region>"
    echo "Example: $0 123456789012 us-east-1"
    exit 1
fi

AWS_ACCOUNT_ID=$1
AWS_REGION=$2

# Login to ECR
echo "Logging in to ECR..."
aws ecr get-login-password --region $AWS_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Function to build and push a service
build_and_push() {
    local service_name=$1
    local service_path=$2
    local image_tag=$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$service_name:latest
    
    echo "Building $service_name..."
    cd $service_path
    docker build --platform linux/amd64 -t $image_tag .
    
    echo "Pushing $service_name..."
    docker push $image_tag
    
    echo "✅ $service_name pushed successfully"
    cd - > /dev/null
}

# Build and push all services
echo "🚀 Starting build and push process..."

build_and_push "product-service" "services/product-service"
build_and_push "cart-vulnerable" "services/cart-service/vulnerable"
build_and_push "cart-fixed" "services/cart-service/fixed"

echo "🎉 All images built and pushed successfully!"
echo ""
echo "Images pushed:"
echo "- $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/product-service:latest"
echo "- $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/cart-vulnerable:latest"
echo "- $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/cart-fixed:latest"
