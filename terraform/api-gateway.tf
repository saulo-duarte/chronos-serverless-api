resource "aws_apigatewayv2_api" "lambda_api_v2" {
  name          = "${var.lambda_function_name}-api-v2"
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "lambda_integration_v2" {
  api_id             = aws_apigatewayv2_api.lambda_api_v2.id
  integration_type   = "AWS_PROXY"
  integration_uri    = aws_lambda_function.go_lambda.invoke_arn
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "lambda_route_proxy" {
  api_id    = aws_apigatewayv2_api.lambda_api_v2.id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.lambda_integration_v2.id}"
}

resource "aws_apigatewayv2_stage" "lambda_stage_v2" {
  api_id      = aws_apigatewayv2_api.lambda_api_v2.id
  name        = "prod"
  auto_deploy = true
}

resource "aws_lambda_permission" "api_gateway_v2" {
  statement_id  = "AllowExecutionFromAPIGatewayV2"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.go_lambda.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_apigatewayv2_api.lambda_api_v2.execution_arn}/*"
}

output "lambda_api_v2_endpoint" {
  value = aws_apigatewayv2_api.lambda_api_v2.api_endpoint
}