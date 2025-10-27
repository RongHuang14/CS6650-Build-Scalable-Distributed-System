output "alb_dns" {
  value = aws_lb.main.dns_name
}

output "sns_topic_arn" {
  value = aws_sns_topic.orders.arn
}

output "sqs_queue_url" {
  value = aws_sqs_queue.orders.url
}

output "ecr_api_url" {
  value = aws_ecr_repository.api.repository_url
}

output "ecr_processor_url" {
  value = aws_ecr_repository.processor.repository_url
}