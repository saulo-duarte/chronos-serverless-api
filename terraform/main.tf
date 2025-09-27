terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.13.1"
    }
  }

  backend "s3" {
    bucket = var.terraform_backend_bucket
    key    = "terraform.tfstate"
    region = var.region
  }
}

provider "aws" {
  region = var.region
}

data "aws_caller_identity" "current" {}

output "lambda_function_name" {
  value = aws_lambda_function.go_lambda.function_name
}

output "lambda_api_custom_endpoint" {
  description = "Novo endpoint da API (https://api.chronosapp.site/) que deve ser usado no seu frontend."
  value       = "https://${aws_apigatewayv2_domain_name.api_custom_domain.domain_name}"
}

output "acm_validation_cname" {
  description = "CNAME record name e value para validar o certificado ACM (Inserir na Hostinger)."
  value = {
    name  = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_name
    type  = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_type
    value = aws_acm_certificate.api_cert.domain_validation_options[0].resource_record_value
  }
}

output "api_gateway_target_cname" {
  description = "CNAME Target Domain (Value) que deve ser inserido na Hostinger (Host: api)."
  value       = aws_apigatewayv2_domain_name.api_custom_domain.domain_name_configuration[0].target_domain_name
}
