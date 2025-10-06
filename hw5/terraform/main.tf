# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
}

# Build & push the Python app image into ECR
# Using null_resource with local-exec provisioner for better macOS compatibility
resource "null_resource" "docker_build_push" {
  triggers = {
    # Trigger rebuild when source files change
    dockerfile_hash = filemd5("../src/Dockerfile")
    main_py_hash    = filemd5("../src/main.py")
    requirements_hash = filemd5("../src/requirements.txt")
  }

  provisioner "local-exec" {
    command = <<-EOT
      # Login to ECR
      aws ecr get-login-password --region ${var.aws_region} | docker login --username AWS --password-stdin ${module.ecr.repository_url}
      
      # Build image for linux/amd64 platform
      docker buildx build --platform linux/amd64 -t ${module.ecr.repository_url}:latest ../src
      
      # Push to ECR
      docker push ${module.ecr.repository_url}:latest
    EOT
  }

  depends_on = [module.ecr]
}
