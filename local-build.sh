#!/bin/bash

# --- Configurações ---
# O nome do binário para o runtime customizado da AWS Lambda
LAMBDA_BINARY="bootstrap"
# O nome do binário para desenvolvimento local
LOCAL_BINARY="chronos-local"

# Garante que o script pare em qualquer erro
set -e

echo "--- Passo 1: Limpeza e Preparação ---"
# Remove binários anteriores e o cache local do SAM para garantir um build limpo
rm -f $LAMBDA_BINARY $LOCAL_BINARY
rm -rf .aws-sam

# --- Lógica Condicional de Execução ---

if [ "$1" == "local" ]; then
    # --- MODO DE DESENVOLVIMENTO LOCAL ---
    echo "--- MODO LOCAL DETECTADO ---"

    echo "Passo 2: Construindo binário local '$LOCAL_BINARY' para $(go env GOOS)/$(go env GOARCH)..."
    # Compila para o sistema operacional e arquitetura local
    go build -o $LOCAL_BINARY main.go

    echo "Binário criado com sucesso:"
    ls -la $LOCAL_BINARY

    echo "Passo 3: Carregando variáveis de ambiente do .env e iniciando o servidor local..."
    
    # Define a porta 3000 para o modo local, igual ao SAM CLI
    LOCAL_PORT="3000"

    if [ -f .env ]; then
        # set -a: marca todas as variáveis definidas a partir daqui como exportáveis.
        set -a
        source .env
        set +a # Volta o comportamento padrão
        
        # Define RUN_MODE e PORT explicitamente no ambiente de exportação
        export RUN_MODE="local" 
        export PORT=$LOCAL_PORT
        echo "Arquivo .env carregado e variáveis exportadas com sucesso."
    else
        echo "Aviso: Arquivo .env não encontrado. Certifique-se de que as variáveis de ambiente estão definidas manualmente."
        export RUN_MODE="local"
        export PORT=$LOCAL_PORT
    fi
    
    echo "Iniciando servidor HTTP local (RUN_MODE=local) em http://localhost:$LOCAL_PORT"
    
    # Executa o binário Go diretamente (as variáveis já estão no ambiente do shell)
    ./$LOCAL_BINARY

elif [ "$1" == "sam" ] || [ -z "$1" ]; then
    # --- MODO SIMULAÇÃO LAMBDA (SAM CLI) ---
    echo "--- MODO SIMULAÇÃO LAMBDA (SAM CLI) DETECTADO ---"
    
    echo "Passo 2: Construindo o binário Go para Linux (nomeado '$LAMBDA_BINARY')..."
    # Compila para o ambiente de execução da AWS Lambda (Linux/AMD64)
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $LAMBDA_BINARY main.go

    echo "Binário criado com sucesso:"
    ls -la $LAMBDA_BINARY

    echo "Passo 3: Iniciando a API local do SAM (http://localhost:3000)."
    # O SAM CLI irá usar o binário 'bootstrap' e ler as variáveis do env.json
    sam local start-api --env-vars env.json --port 3000

else
    echo "Uso inválido."
    echo "Para rodar localmente: ./run.sh local"
    echo "Para simular Lambda com SAM: ./run.sh sam"
    exit 1
fi