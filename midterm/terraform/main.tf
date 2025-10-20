# terraform/main.tf

# Provider Configuration
terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# Data Sources
data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
  
  filter {
    name   = "availability-zone"
    values = var.availability_zones
  }
}

# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = var.ecs_cluster_name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = var.tags
}

# CloudWatch Log Groups
resource "aws_cloudwatch_log_group" "product_service" {
  name              = "/ecs/${var.ecs_cluster_name}/product-service"
  retention_in_days = 7
  tags              = var.tags
}

resource "aws_cloudwatch_log_group" "cart_vulnerable" {
  name              = "/ecs/${var.ecs_cluster_name}/cart-vulnerable"
  retention_in_days = 7
  tags              = var.tags
}

resource "aws_cloudwatch_log_group" "cart_fixed" {
  name              = "/ecs/${var.ecs_cluster_name}/cart-fixed"
  retention_in_days = 7
  tags              = var.tags
}

# Security Groups
resource "aws_security_group" "alb" {
  name        = "${var.ecs_cluster_name}-alb-sg"
  description = "Security group for ALB"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

resource "aws_security_group" "ecs_tasks" {
  name        = "${var.ecs_cluster_name}-tasks-sg"
  description = "Security group for ECS tasks"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    from_port       = 8000
    to_port         = 8000
    protocol        = "tcp"
    security_groups = [aws_security_group.alb.id]
  }

  # Allow inter-service communication
  ingress {
    from_port = 8000
    to_port   = 8000
    protocol  = "tcp"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = var.tags
}

# Use existing AWS Academy Lab Role instead of creating new ones
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Application Load Balancer
resource "aws_lb" "main" {
  name               = "${var.ecs_cluster_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets           = data.aws_subnets.default.ids

  enable_deletion_protection = false
  enable_http2              = true

  tags = var.tags
}

# Target Groups
resource "aws_lb_target_group" "product_service" {
  name        = "cb-demo-product-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = var.tags
}

resource "aws_lb_target_group" "cart_vulnerable" {
  name        = "cb-demo-cart-vuln-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = var.tags
}

resource "aws_lb_target_group" "cart_fixed" {
  name        = "cb-demo-cart-fixed-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = data.aws_vpc.default.id
  target_type = "ip"

  health_check {
    enabled             = true
    healthy_threshold   = 2
    unhealthy_threshold = 2
    timeout             = 5
    interval            = 30
    path                = "/health"
    matcher             = "200"
  }

  deregistration_delay = 30

  tags = var.tags
}

# ALB Listener
resource "aws_lb_listener" "main" {
  load_balancer_arn = aws_lb.main.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type = "fixed-response"
    fixed_response {
      content_type = "text/plain"
      message_body = "Circuit Breaker Demo - Path not found"
      status_code  = "404"
    }
  }
}

# ALB Listener Rules
resource "aws_lb_listener_rule" "product_service" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 100

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.product_service.arn
  }

  condition {
    path_pattern {
      values = [
        "/products",
        "/products/*",
        "/health",
        "/fail/*",
        "/latency"
      ]
    }
  }
}

resource "aws_lb_listener_rule" "product_service_extra" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 140

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.product_service.arn
  }

  condition {
    path_pattern {
      values = [
        "/reset"
      ]
    }
  }
}

## Optional: Redirect legacy paths to current control endpoints
resource "aws_lb_listener_rule" "product_service_control_redirects" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 150

  action {
    type = "redirect"
    redirect {
      host        = "#{host}"
      path        = "/fail/on"
      port        = "80"
      protocol    = "HTTP"
      query       = "#{query}"
      status_code = "HTTP_301"
    }
  }

  condition {
    path_pattern {
      values = ["/crash"]
    }
  }
}

resource "aws_lb_listener_rule" "product_service_control_redirects_recover" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 160

  action {
    type = "redirect"
    redirect {
      host        = "#{host}"
      path        = "/fail/off"
      port        = "80"
      protocol    = "HTTP"
      query       = "#{query}"
      status_code = "HTTP_301"
    }
  }

  condition {
    path_pattern {
      values = ["/recover"]
    }
  }
}

resource "aws_lb_listener_rule" "cart_vulnerable" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 200

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cart_vulnerable.arn
  }

  condition {
    path_pattern {
      values = ["/vulnerable/*"]
    }
  }
}

resource "aws_lb_listener_rule" "cart_fixed" {
  listener_arn = aws_lb_listener.main.arn
  priority     = 300

  action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.cart_fixed.arn
  }

  condition {
    path_pattern {
      values = ["/fixed/*"]
    }
  }
}

# ECS Task Definitions
resource "aws_ecs_task_definition" "product_service" {
  family                   = "product-service"
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = var.task_cpu
  memory                  = var.task_memory
  execution_role_arn      = data.aws_iam_role.lab_role.arn
  task_role_arn           = data.aws_iam_role.lab_role.arn

  container_definitions = jsonencode([
    {
      name  = "product-service"
      image = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/product-service:latest"
      
      portMappings = [
        {
          containerPort = var.product_service_port
        }
      ]
      
      environment = [
        {
          name  = "FAILURE_MODE"
          value = "false"
        },
        {
          name  = "FAILURE_RATE"
          value = "0.0"
        },
        {
          name  = "LATENCY_MS"
          value = "0"
        }
      ]
      
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.product_service.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

resource "aws_ecs_task_definition" "cart_vulnerable" {
  family                   = "cart-vulnerable"
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = var.task_cpu
  memory                  = var.task_memory
  execution_role_arn      = data.aws_iam_role.lab_role.arn
  task_role_arn           = data.aws_iam_role.lab_role.arn
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([
    {
      name  = "cart-vulnerable"
      image = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/cart-vulnerable:latest"
      
      portMappings = [
        {
          containerPort = var.cart_service_port
        }
      ]
      
      environment = [
        {
          name  = "PRODUCT_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}"
        }
      ]
      
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.cart_vulnerable.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

resource "aws_ecs_task_definition" "cart_fixed" {
  family                   = "cart-fixed"
  network_mode            = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                     = var.task_cpu
  memory                  = var.task_memory
  execution_role_arn      = data.aws_iam_role.lab_role.arn
  task_role_arn           = data.aws_iam_role.lab_role.arn
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([
    {
      name  = "cart-fixed"
      image = "${var.aws_account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/cart-fixed:latest"
      
      portMappings = [
        {
          containerPort = var.cart_service_port
        }
      ]
      
      environment = [
        {
          name  = "PRODUCT_SERVICE_URL"
          value = "http://${aws_lb.main.dns_name}"
        }
      ]
      
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.cart_fixed.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ecs"
        }
      }
    }
  ])

  tags = var.tags
}

# ECS Services
resource "aws_ecs_service" "product_service" {
  name            = "product-service"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.product_service.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.product_service.arn
    container_name   = "product-service"
    container_port   = var.product_service_port
  }

  depends_on = [aws_lb_listener.main]

  tags = var.tags
}

resource "aws_ecs_service" "cart_vulnerable" {
  name            = "cart-vulnerable"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.cart_vulnerable.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.cart_vulnerable.arn
    container_name   = "cart-vulnerable"
    container_port   = var.cart_service_port
  }

  depends_on = [
    aws_lb_listener.main,
    aws_ecs_service.product_service
  ]

  tags = var.tags
}

resource "aws_ecs_service" "cart_fixed" {
  name            = "cart-fixed"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.cart_fixed.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = data.aws_subnets.default.ids
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.cart_fixed.arn
    container_name   = "cart-fixed"
    container_port   = var.cart_service_port
  }

  depends_on = [
    aws_lb_listener.main,
    aws_ecs_service.product_service
  ]

  tags = var.tags
}

# Outputs
output "alb_url" {
  description = "URL of the Application Load Balancer"
  value       = "http://${aws_lb.main.dns_name}"
}

output "service_endpoints" {
  description = "Service endpoints for testing"
  value = {
    product_service  = "http://${aws_lb.main.dns_name}/products"
    cart_vulnerable  = "http://${aws_lb.main.dns_name}/vulnerable/cart"
    cart_fixed      = "http://${aws_lb.main.dns_name}/fixed/cart"
    trigger_crash   = "http://${aws_lb.main.dns_name}/fail/on"
    trigger_recover = "http://${aws_lb.main.dns_name}/fail/off"
  }
}

output "cloudwatch_logs" {
  description = "CloudWatch log groups for monitoring"
  value = {
    product_service = aws_cloudwatch_log_group.product_service.name
    cart_vulnerable = aws_cloudwatch_log_group.cart_vulnerable.name
    cart_fixed     = aws_cloudwatch_log_group.cart_fixed.name
  }
}