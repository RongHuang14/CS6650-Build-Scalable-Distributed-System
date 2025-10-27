# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "${var.project_name}-cluster"
}

# ECR Repositories
resource "aws_ecr_repository" "api" {
  name = "${var.project_name}-api"
}

resource "aws_ecr_repository" "processor" {
  name = "${var.project_name}-processor"
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "api" {
  name              = "/ecs/${var.project_name}-api"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "processor" {
  name              = "/ecs/${var.project_name}-processor"
  retention_in_days = 7
}

# Task Definition - API
resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project_name}-api"
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = "256"
  memory                  = "512"
  execution_role_arn      = aws_iam_role.ecs_execution_role.arn
  task_role_arn          = aws_iam_role.ecs_task_role.arn
  
  container_definitions = jsonencode([{
    name  = "api"
    image = "${aws_ecr_repository.api.repository_url}:latest"
    
    environment = [
      {
        name  = "SNS_TOPIC_ARN"
        value = aws_sns_topic.orders.arn
      },
      {
        name  = "AWS_REGION"
        value = var.aws_region
      }
    ]
    
    portMappings = [{
      containerPort = 8080
    }]
    
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.api.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

# Task Definition - Processor
resource "aws_ecs_task_definition" "processor" {
  family                   = "${var.project_name}-processor"
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = "256"
  memory                  = "512"
  execution_role_arn      = aws_iam_role.ecs_execution_role.arn
  task_role_arn          = aws_iam_role.ecs_task_role.arn
  
  container_definitions = jsonencode([{
    name  = "processor"
    image = "${aws_ecr_repository.processor.repository_url}:latest"
    
    environment = [
      {
        name  = "SQS_QUEUE_URL"
        value = aws_sqs_queue.orders.url
      },
      {
        name  = "AWS_REGION"
        value = var.aws_region
      },
      {
        name  = "WORKER_COUNT"
        value = tostring(var.worker_count)
      }
    ]
    
    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.processor.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

# ECS Service - API
resource "aws_ecs_service" "api" {
  name            = "${var.project_name}-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = 1
  launch_type     = "FARGATE"
  
  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }
  
  load_balancer {
    target_group_arn = aws_lb_target_group.api.arn
    container_name   = "api"
    container_port   = 8080
  }
}

# ECS Service - Processor
resource "aws_ecs_service" "processor" {
  name            = "${var.project_name}-processor"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.processor.arn
  desired_count   = 1  # Always 1 task; scale workers within the task via WORKER_COUNT env var
  launch_type     = "FARGATE"
  
  network_configuration {
    subnets         = aws_subnet.private[*].id
    security_groups = [aws_security_group.ecs.id]
  }
}