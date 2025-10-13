output "ecs_cluster_name" {
  description = "Name of the ECS cluster"
  value       = aws_ecs_cluster.main.name
}

output "ecs_service_name" {
  description = "Name of the ECS service"
  value       = aws_ecs_service.app.name
}

output "task_public_ip" {
  description = "Public IP of the ECS task (check ECS Console for actual IP)"
  value       = "Check AWS ECS Console > Cluster > Service > Tasks for the public IP"
}

output "container_port" {
  description = "Port the container is listening on"
  value       = var.container_port
}

output "cloudwatch_log_group" {
  description = "CloudWatch Log Group for application logs"
  value       = aws_cloudwatch_log_group.app.name
}

output "instructions" {
  description = "Next steps"
  value       = <<-EOT
    
    ========================================
    Part 2 Infrastructure Deployed! 
    ========================================
    
    Single Instance Configuration:
    - CPU: ${var.task_cpu} units (0.25 vCPU)
    - Memory: ${var.task_memory} MB
    - Instances: 1
    
    Next Steps:
    
    1. Get the Public IP:
       aws ecs list-tasks --cluster ${aws_ecs_cluster.main.name} --service-name ${aws_ecs_service.app.name}
       aws ecs describe-tasks --cluster ${aws_ecs_cluster.main.name} --tasks <TASK_ARN>
       
       Or check in AWS Console:
       ECS > Clusters > ${aws_ecs_cluster.main.name} > Services > ${aws_ecs_service.app.name} > Tasks
    
    2. Test the service:
       curl http://<PUBLIC_IP>:8080/health
       curl "http://<PUBLIC_IP>:8080/products/search?q=Electronics"
    
    3. Run Locust load tests:
       - Test 1: 5 users for 2 minutes
       - Test 2: 20 users for 3 minutes
    
    4. Monitor in CloudWatch:
       - CPU Utilization
       - Memory Utilization
       - Response times
    
    Goal: Find the breaking point!
    ========================================
  EOT
}

