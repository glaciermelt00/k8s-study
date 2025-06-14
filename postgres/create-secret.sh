#!/bin/bash

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# .envrcの環境変数を読み込む
if [ -f "$PROJECT_ROOT/.envrc" ]; then
    source "$PROJECT_ROOT/.envrc"
else
    echo "Error: .envrc file not found at $PROJECT_ROOT/.envrc"
    exit 1
fi

# 環境変数が設定されているか確認
if [ -z "$POSTGRES_USER" ] || [ -z "$POSTGRES_PASSWORD" ] || [ -z "$POSTGRES_DB" ]; then
    echo "Error: Required environment variables are not set"
    echo "Please check your .envrc file"
    exit 1
fi

# Secret YAMLマニフェストを生成
kubectl create secret generic postgres-secret \
    --namespace=default \
    --from-literal=POSTGRES_USER="$POSTGRES_USER" \
    --from-literal=POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    --from-literal=POSTGRES_DB="$POSTGRES_DB" \
    --dry-run=client -o yaml > "$SCRIPT_DIR/secret.yaml"

echo "Secret manifest created: $SCRIPT_DIR/secret.yaml"
echo ""
echo "To apply the secret:"
echo "  kubectl apply -f $SCRIPT_DIR/secret.yaml"
