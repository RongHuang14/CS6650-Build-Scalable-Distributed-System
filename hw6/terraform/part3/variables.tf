variable "aws_region" {
  description = "AWS region for deployment"
  type        = string
  default     = "us-west-2"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "hw6-part3"
}

variable "min_capacity" {
  description = "Minimum number of ECS tasks"
  type        = number
  default     = 2
}

variable "max_capacity" {
  description = "Maximum number of ECS tasks"
  type        = number
  default     = 4
}

variable "cpu_target_value" {
  description = "Target CPU utilization percentage for auto scaling"
  type        = number
  default     = 70
}

variable "scale_in_cooldown" {
  description = "Cooldown period (seconds) before allowing scale-in"
  type        = number
  default     = 300
}

variable "scale_out_cooldown" {
  description = "Cooldown period (seconds) before allowing scale-out"
  type        = number
  default     = 300
}

