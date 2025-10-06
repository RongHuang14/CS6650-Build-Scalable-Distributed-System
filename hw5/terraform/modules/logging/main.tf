# Create CloudWatch log group
resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.service_name}"
  retention_in_days = var.retention_in_days

  tags = {
    Name = "${var.service_name}-logs"
  }
}
