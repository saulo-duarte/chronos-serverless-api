resource "aws_acm_certificate" "api_cert" {
  provider                  = aws.api_region
  domain_name               = local.api_domain_name
  validation_method         = "DNS"
  subject_alternative_names = [var.root_domain]

  tags = {
    Name = "${var.lambda_function_name}-api-cert"
  }

  lifecycle {
    create_before_destroy = true
  }
}

resource "aws_acm_certificate_validation" "api_cert_validation" {
  provider                = aws.api_region
  certificate_arn         = aws_acm_certificate.api_cert.arn
  validation_record_fqdns = [] 
}
