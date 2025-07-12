#!/bin/bash

# 色付きの出力用
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== PostgreSQL接続スクリプト ===${NC}\n"

# 1. 既存のポートフォワードを終了
echo -e "${YELLOW}既存のポートフォワードを確認中...${NC}"
if pgrep -f "kubectl port-forward postgres-0" > /dev/null; then
    echo "既存のポートフォワードを終了します..."
    pkill -f "kubectl port-forward postgres-0"
    sleep 2
fi

# 2. PostgreSQLポッドの状態確認
echo -e "${YELLOW}PostgreSQLポッドの状態を確認中...${NC}"
POD_STATUS=$(kubectl get pod postgres-0 -o jsonpath='{.status.phase}' 2>/dev/null)
if [ "$POD_STATUS" != "Running" ]; then
    echo -e "${RED}エラー: PostgreSQLポッドが実行中ではありません (状態: $POD_STATUS)${NC}"
    exit 1
fi
echo -e "${GREEN}✓ PostgreSQLポッドが実行中です${NC}"

# 3. パスワードを取得
echo -e "\n${YELLOW}パスワードを取得中...${NC}"
PGPASSWORD=$(kubectl get secret postgres-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d)
PGUSER=$(kubectl get secret postgres-secret -o jsonpath='{.data.POSTGRES_USER}' | base64 -d)
PGDATABASE=$(kubectl get secret postgres-secret -o jsonpath='{.data.POSTGRES_DB}' | base64 -d)

if [ -z "$PGPASSWORD" ]; then
    echo -e "${RED}エラー: パスワードを取得できませんでした${NC}"
    exit 1
fi

# 4. ポートフォワードを開始
echo -e "\n${YELLOW}ポートフォワードを開始します...${NC}"
kubectl port-forward postgres-0 15432:5432 &
PF_PID=$!
trap 'kill $PF_PID' EXIT
sleep 3

# ポートフォワードが成功したか確認
if ! ps -p $PF_PID > /dev/null; then
    echo -e "${RED}エラー: ポートフォワードの開始に失敗しました${NC}"
    exit 1
fi
echo -e "${GREEN}✓ ポートフォワードが開始されました (PID: $PF_PID)${NC}"

# 5. 接続情報を表示
echo -e "\n${BLUE}=== 接続情報 ===${NC}"
echo -e "ホスト: localhost"
echo -e "ポート: 15432"
echo -e "ユーザー: $PGUSER"
echo -e "データベース: $PGDATABASE"
echo -e "パスワード: ${YELLOW}[自動的に設定されます]${NC}"

# 6. psqlで接続
echo -e "\n${GREEN}PostgreSQLに接続します...${NC}"
echo -e "${YELLOW}終了するには \\q を入力してください${NC}\n"

# パスワードを環境変数として設定して接続
PGPASSWORD=$PGPASSWORD psql -h localhost -p 15432 -U $PGUSER -d $PGDATABASE

# 7. クリーンアップ
echo -e "\n${YELLOW}ポートフォワードを終了します...${NC}"
kill $PF_PID 2>/dev/null
echo -e "${GREEN}接続を終了しました${NC}"