variable "project_name" {
  default = "hw7"
}

variable "aws_region" {
  default = "us-west-2"
}

variable "worker_count" {
  description = "Number of concurrent goroutines within the processor task (for Phase 5 testing)"
  default     = 5  # Phase 5: test with 1, 5, 20, 100
}