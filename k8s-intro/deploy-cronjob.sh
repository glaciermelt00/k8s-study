#!/bin/bash
set -euo pipefail

echo "🚀 CronJob のデプロイを開始します..."

# Docker イメージのビルド
echo "📦 Slack メトリクスバッチのイメージをビルドしています..."
cd batch-scripts
docker build -t slack-metrics-batch:latest .
cd ..

# Minikubeにイメージをロード
echo "📤 Minikubeにイメージをロードしています..."
minikube image load slack-metrics-batch:latest

# ConfigMap のデプロイ
echo "📋 ConfigMap をデプロイしています..."
kubectl apply -f deployments/cronjob/slack-metrics-configmap.yaml

# Secret のデプロイ（環境変数から Webhook URL を取得）
echo "🔐 Slack Secret をデプロイしています..."
if [ -z "${SLACK_WEBHOOK_URL:-}" ]; then
    echo "❌ エラー: SLACK_WEBHOOK_URL 環境変数が設定されていません"
    echo "   以下のコマンドで環境変数を設定してください:"
    echo "   export SLACK_WEBHOOK_URL='https://hooks.slack.com/services/YOUR/WEBHOOK/URL'"
    exit 1
fi

# 環境変数を使用して Secret を作成
kubectl create secret generic slack-secret \
    --from-literal=webhook-url="$SLACK_WEBHOOK_URL" \
    --dry-run=client -o yaml | kubectl apply -f -

# CronJob のデプロイ
echo "⏰ CronJob をデプロイしています..."
kubectl apply -f deployments/cronjob/slack-metrics-cronjob.yaml

echo "✅ CronJob のデプロイが完了しました！"
echo ""
echo "📊 デプロイされたリソースの確認:"
kubectl get configmap slack-metrics-config
kubectl get secret slack-secret
kubectl get cronjob slack-metrics-cronjob

echo ""
echo "🔍 CronJob の詳細:"
kubectl describe cronjob slack-metrics-cronjob

echo ""
echo "💡 手動でジョブを実行するには:"
echo "kubectl create job --from=cronjob/slack-metrics-cronjob manual-test-$(date +%s)"