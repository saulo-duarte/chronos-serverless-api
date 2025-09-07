#!/bin/bash

OUTPUT_BINARY="bootstrap"
LAMBDA_FUNCTION_NAME="ChronosFunction"
BUILD_DIR=".aws-sam/build/$LAMBDA_FUNCTION_NAME"

set -e

echo "Passo 1: Construindo o binário Go para Linux..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $OUTPUT_BINARY main.go

echo "Binário criado com sucesso:"
ls -la $OUTPUT_BINARY

echo "Passo 2: Movendo o binário para o diretório de build do SAM..."
mkdir -p $BUILD_DIR

mv $OUTPUT_BINARY $BUILD_DIR/

echo "Passo 3: Iniciando a API local do SAM..."
sam local start-api --env-vars env.json --port 3000