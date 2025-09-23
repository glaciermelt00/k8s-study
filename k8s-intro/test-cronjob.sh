#!/bin/bash
set -euo pipefail

echo "🧪 CronJob のテストを開始します..."

# Docker イメージのビルドとMinikubeへのロード
echo "📦 Docker イメージをビルドしています..."
cd batch-scripts
docker build -t slack-metrics-batch:latest .
cd ..

# Minikubeにイメージをロード
echo "📤 Minikubeにイメージをロードしています..."
minikube image load slack-metrics-batch:latest

# 手動でジョブを作成してテスト
JOB_NAME="manual-test-$(date +%s)"
echo "📝 テスト用ジョブ名: $JOB_NAME"

# CronJob から手動でジョブを作成
echo "🚀 CronJob から手動でジョブを作成しています..."
kubectl create job --from=cronjob/slack-metrics-cronjob "$JOB_NAME"

# ジョブの状態を監視
echo "⏳ ジョブの実行を待っています..."
kubectl wait --for=condition=complete --timeout=120s job/"$JOB_NAME" || {
    echo "❌ ジョブがタイムアウトしました"
    echo "📋 ジョブの状態:"
    kubectl describe job "$JOB_NAME"
    echo "📄 Pod の状態:"
    kubectl get pods -l job-name="$JOB_NAME" -o wide
    echo "📄 Pod のログ:"
    kubectl logs -l job-name="$JOB_NAME" --tail=50 || echo "ログの取得に失敗しました"
    
    # Pod のイベントを確認
    POD_NAME=$(kubectl get pods -l job-name="$JOB_NAME" -o jsonpath='{.items[0].metadata.name}')
    if [ -n "$POD_NAME" ]; then
        echo "📋 Pod のイベント:"
        kubectl describe pod "$POD_NAME" | grep -A 10 "Events:" || true
    fi
    exit 1
}

echo "✅ ジョブが正常に完了しました！"

# ログの確認
echo "📄 ジョブのログ:"
kubectl logs -l job-name="$JOB_NAME"

# クリーンアップ
echo "🧹 テスト用ジョブをクリーンアップしています..."
kubectl delete job "$JOB_NAME"

echo "✨ テストが完了しました！"