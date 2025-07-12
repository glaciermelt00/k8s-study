#!/bin/bash
set -euo pipefail

# エラーハンドリング
trap 'echo -e "${RED}エラーが発生しました。終了します。${NC}" >&2' ERR

# 色付きの出力用
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=== マイグレーション設定のテストスクリプト ==="

# 1. 環境変数が設定されているか確認
echo -e "\n${YELLOW}1. 環境変数の確認${NC}"
if [ -z "$POSTGRES_PASSWORD" ]; then
    echo -e "${RED}エラー: POSTGRES_PASSWORD が設定されていません${NC}"
    echo "先に 'source .envrc' を実行してください"
    exit 1
fi
echo -e "${GREEN}✓ 環境変数が設定されています${NC}"

# 2. PostgreSQLのSecretが存在するか確認
echo -e "\n${YELLOW}2. PostgreSQL Secretの確認${NC}"
if ! kubectl get secret postgres-secret > /dev/null 2>&1; then
    echo -e "${RED}エラー: postgres-secret が見つかりません${NC}"
    echo "先に 'scripts/create-secret.sh' を実行してください"
    exit 1
fi
echo -e "${GREEN}✓ postgres-secret が存在します${NC}"

# 3. PostgreSQL StatefulSetが動作しているか確認
echo -e "\n${YELLOW}3. PostgreSQL StatefulSetの確認${NC}"
PG_STATUS=$(kubectl get statefulset postgres -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo "0")
if [ "${PG_STATUS}" != "1" ]; then
    echo -e "${RED}エラー: PostgreSQL StatefulSetが準備できていません${NC}"
    echo "先に PostgreSQL をデプロイしてください"
    exit 1
fi
echo -e "${GREEN}✓ PostgreSQL StatefulSetが動作しています${NC}"

# 4. マイグレーションAPIのビルド
echo -e "\n${YELLOW}4. マイグレーションAPIのビルド${NC}"
if ! docker build -t migration-api:latest -f apps/migration/Dockerfile apps/migration/; then
    echo -e "${RED}エラー: Dockerイメージのビルドに失敗しました${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Dockerイメージのビルドが完了しました${NC}"

# 5. 既存のジョブを削除（存在する場合）
echo -e "\n${YELLOW}5. 既存のジョブを削除${NC}"
kubectl delete job migration-job --ignore-not-found=true

# 6. マイグレーションジョブのデプロイ
echo -e "\n${YELLOW}6. マイグレーションジョブのデプロイ${NC}"
if ! kubectl apply -f deployments/migration/job.yaml; then
    echo -e "${RED}エラー: ジョブのデプロイに失敗しました${NC}"
    exit 1
fi
echo -e "${GREEN}✓ ジョブをデプロイしました${NC}"

# 7. ジョブの完了を待つ
echo -e "\n${YELLOW}7. ジョブの実行を監視${NC}"
echo "ジョブの完了を待っています..."

# 最大60秒待つ
for i in {1..60}; do
    STATUS=$(kubectl get job migration-job -o jsonpath='{.status.conditions[?(@.type=="Complete")].status}' 2>/dev/null || echo "")
    FAILED=$(kubectl get job migration-job -o jsonpath='{.status.conditions[?(@.type=="Failed")].status}' 2>/dev/null || echo "")
    
    if [ "${STATUS}" = "True" ]; then
        echo -e "\n${GREEN}✓ ジョブが正常に完了しました${NC}"
        break
    elif [ "${FAILED}" = "True" ]; then
        echo -e "\n${RED}✗ ジョブが失敗しました${NC}"
        echo -e "\n${YELLOW}ログ:${NC}"
        kubectl logs job/migration-job || echo "ログの取得に失敗しました"
        exit 1
    fi
    
    echo -n "."
    sleep 1
done

# タイムアウトチェック
if [ $i -eq 60 ]; then
    echo -e "\n${RED}タイムアウト: ジョブが60秒以内に完了しませんでした${NC}"
    kubectl describe job migration-job
    exit 1
fi

# 8. ジョブのログを表示
echo -e "\n${YELLOW}8. ジョブのログ${NC}"
kubectl logs job/migration-job

# 9. データベースにテーブルが作成されたか確認
echo -e "\n${YELLOW}9. データベースの確認${NC}"
kubectl exec -i postgres-0 -- psql -U postgres -d postgres -c "\dt"

echo -e "\n${GREEN}=== テスト完了 ===${NC}"