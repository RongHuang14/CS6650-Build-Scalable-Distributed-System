variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "hw6-part2"
}

variable "container_name" {
  description = "Name of the container"
  type        = string
  default     = "product-search"
}

variable "container_image" {
  description = "Docker image to deploy (ECR URI or Docker Hub)"
  type        = string
  # Replace with your ECR URI
  # Example: 123456789012.dkr.ecr.us-west-2.amazonaws.com/hw6-product-search:latest
}

variable "container_port" {
  description = "Port the container listens on"
  type        = number
  default     = 8080
}

variable "task_cpu" {
  description = "CPU units for the task (256 = 0.25 vCPU)"
  type        = string
  default     = "256"
}

variable "task_memory" {
  description = "Memory for the task in MB"
  type        = string
  default     = "512"
}

