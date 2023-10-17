provider "aws" {
  region = "us-east-1"
}

resource "aws_iam_role" "lambda_role" {
  name = "lambda-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "lambda_policy" {
  name = "lambda_policy"
  role = aws_iam_role.lambda_role.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:*",
        "kinesis:*",
        "cloudwatch:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "log_producer" {
  function_name = "log-producer"
  handler       = "index.handler"
  runtime       = "nodejs18.x"
  role          = aws_iam_role.lambda_role.arn

  filename = "../dist/log-producer.zip" # Update the path to your zip file
}

resource "aws_kinesis_stream" "api_logs_to_cloudwatch" {
  name        = "api_logs_to_cloudwatch"
  shard_count = 1
}

resource "aws_cloudwatch_log_group" "log_group" {
  name = "/aws/lambda/log-producer"
}

resource "aws_cloudwatch_log_subscription_filter" "subscription_filter" {
  name            = "clickhouse-api-log"
  log_group_name  = aws_cloudwatch_log_group.log_group.name
  filter_pattern  = "{ $.tenantId=* }"
  destination_arn = aws_kinesis_stream.api_logs_to_cloudwatch.arn

  role_arn = aws_iam_role.lambda_role.arn # You might want to create a separate role for this
}

variable "clickhouse_url" {
  description = "Clickhouse URL to ingest log events"
  type = string
}
  
resource "aws_lambda_event_source_mapping" "kinesis_event_source_mapping" {
  event_source_arn  = aws_kinesis_stream.api_logs_to_cloudwatch.arn
  function_name     = aws_lambda_function.kinesis_consumer.arn
  starting_position = "TRIM_HORIZON"
}

resource "aws_lambda_function" "kinesis_consumer" {
  function_name = "kinesis-consumer"
  handler       = "index.handler"
  runtime       = "nodejs18.x"
  role          = aws_iam_role.lambda_kinesis_role.arn

  filename = "../dist/log-sink.zip"

  environment {
    variables = {
      CLICKHOUSE_URL =  var.clickhouse_url
    }
  }
}
