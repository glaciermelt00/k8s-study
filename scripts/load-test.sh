#!/bin/bash

# HPA の動作を確認するための負荷テストスクリプト

echo "Starting load test for HPA testing..."
echo "This will send multiple concurrent requests to the API"
echo ""

# 色の定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# API のエンドポイント
API_ENDPOINT="${API_ENDPOINT:-http://sm-api.local}"

# 負荷テストの設定
CONCURRENT_USERS="${CONCURRENT_USERS:-50}"
DURATION="${DURATION:-300}"  # 秒
REQUEST_INTERVAL="${REQUEST_INTERVAL:-0.1}"  # 秒

echo -e "${YELLOW}Test Configuration:${NC}"
echo "  API Endpoint: $API_ENDPOINT"
echo "  Concurrent Users: $CONCURRENT_USERS"
echo "  Duration: $DURATION seconds"
echo "  Request Interval: $REQUEST_INTERVAL seconds"
echo ""

# HPA の初期状態を表示
echo -e "${GREEN}Initial HPA status:${NC}"
kubectl get hpa slack-metrics-api-hpa
echo ""

# 負荷テスト用の関数
run_load_test() {
    local user_id=$1
    local end_time=$(($(date +%s) + DURATION))

    while [[ $(date +%s) -lt $end_time ]]; do
        # /metrics エンドポイントにリクエストを送信
        curl -s -o /dev/null -w "User $user_id: %{http_code} - %{time_total}s\n" \
            "$API_ENDPOINT/metrics" &

        # 次のリクエストまで待機
        sleep $REQUEST_INTERVAL
    done
}

# 監視用の関数
monitor_hpa() {
    while true; do
        echo -e "\n${YELLOW}[$(date +"%Y-%m-%d %H:%M:%S")] HPA Status:${NC}"
        kubectl get hpa slack-metrics-api-hpa
        kubectl get pods -l app=slack-metrics-api --no-headers | wc -l | xargs echo "Current pod count:"
        sleep 10
    done
}

# Ctrl+C でクリーンアップ
cleanup() {
    echo -e "\n${RED}Stopping load test...${NC}"
    kill $(jobs -p) 2>/dev/null
    exit 0
}
trap cleanup INT

# 監視を開始（バックグラウンド）
monitor_hpa &
MONITOR_PID=$!

echo -e "${GREEN}Starting load test with $CONCURRENT_USERS concurrent users...${NC}"
echo ""

# 複数のユーザーで同時に負荷テストを実行
for i in $(seq 1 $CONCURRENT_USERS); do
    run_load_test $i &
done

# 負荷テストが終わるまで待機
sleep $DURATION

# クリーンアップ
echo -e "\n${GREEN}Load test completed!${NC}"
kill $MONITOR_PID 2>/dev/null

# 最終的な HPA の状態を表示
echo -e "\n${GREEN}Final HPA status:${NC}"
kubectl get hpa slack-metrics-api-hpa
echo ""
kubectl get pods -l app=slack-metrics-api

echo -e "\n${YELLOW}Note:${NC} It may take a few minutes for the HPA to scale down after the load test."
