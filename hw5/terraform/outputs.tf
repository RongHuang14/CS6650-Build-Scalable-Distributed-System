output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = module.ecr.repository_url
}

output "log_group_name" {
  description = "Name of the CloudWatch log group"
  value       = module.logging.log_group_name
}

output "vpc_id" {
  description = "ID of the VPC"
  value       = module.network.vpc_id
}
