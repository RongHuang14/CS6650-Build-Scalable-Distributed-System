# terraform/variables.tf

# AWS Configuration
variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "aws_account_id" {
  description = "AWS account ID"
  type        = string
  default     = "776059864518"
}

# Network Configuration
variable "availability_zones" {
  description = "Availability zones for deployment"
  type        = list(string)
  default     = ["us-west-2a", "us-west-2b"]
}

# ECS Configuration
variable "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  type        = string
  default     = "circuit-breaker-demo-cluster"
}

variable "task_cpu" {
  description = "CPU units for ECS tasks"
  type        = string
  default     = "256"
}

variable "task_memory" {
  description = "Memory for ECS tasks"
  type        = string
  default     = "512"
}

# Service Configuration
variable "product_service_port" {
  description = "Port for product service"
  type        = number
  default     = 8000
}

variable "cart_service_port" {
  description = "Port for cart services"
  type        = number
  default     = 8000
}

# Tags
variable "tags" {
  description = "Tags for all resources"
  type        = map(string)
  default = {
    Project     = "CircuitBreakerDemo"
    Course      = "CS6650"
    Environment = "Demo"
  }
}