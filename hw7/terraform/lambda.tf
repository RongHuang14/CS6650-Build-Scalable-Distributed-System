# Lambda Function
resource "aws_lambda_function" "processor" {
  filename         = "../lambda/function.zip"
  function_name    = "${var.project_name}-lambda"
  role            = aws_iam_role.lambda_role.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  memory_size     = 512
  timeout         = 10
  
  environment {
    variables = {
      PROCESSING_TIME = "3"
    }
  }
}

# Lambda Permission for SNS
resource "aws_lambda_permission" "sns" {
  statement_id  = "AllowSNSInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.processor.function_name
  principal     = "sns.amazonaws.com"
  source_arn    = aws_sns_topic.orders.arn
}

# SNS to Lambda Subscription
resource "aws_sns_topic_subscription" "lambda" {
  topic_arn = aws_sns_topic.orders.arn
  protocol  = "lambda"
  endpoint  = aws_lambda_function.processor.arn
}

# CloudWatch Log Group for Lambda
resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${var.project_name}-lambda"
  retention_in_days = 7
}