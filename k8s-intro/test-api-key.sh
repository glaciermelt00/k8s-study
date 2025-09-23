#!/bin/bash

# API キーを引数として受け取る
API_KEY="${1:-$CLAUDE_CODE_OAUTH_TOKEN}"

if [ -z "$API_KEY" ]; then
    echo "使用方法: ./test-api-key.sh <API_KEY>"
    echo "または環境変数 CLAUDE_CODE_OAUTH_TOKEN を設定してください"
    exit 1
fi

echo "API キーをテストしています..."

# Claude API にテストリクエストを送信
response=$(curl -s -w "\n%{http_code}" https://api.anthropic.com/v1/messages \
  -H "x-api-key: $API_KEY" \
  -H "anthropic-version: 2023-06-01" \
  -H "content-type: application/json" \
  -d '{
    "model": "claude-3-haiku-20240307",
    "max_tokens": 10,
    "messages": [
      {"role": "user", "content": "Hello"}
    ]
  }')

# HTTPステータスコードを取得
http_code=$(echo "$response" | tail -n1)
body=$(echo "$response" | sed '$d')

# 結果を表示
if [ "$http_code" = "200" ]; then
    echo "✅ API キーは有効です！"
    echo "レスポンス: $body"
elif [ "$http_code" = "401" ]; then
    echo "❌ API キーが無効です"
    echo "エラー: $body"
elif [ "$http_code" = "403" ]; then
    echo "❌ API キーのアクセス権限がありません"
    echo "エラー: $body"
else
    echo "❌ エラーが発生しました (HTTP $http_code)"
    echo "詳細: $body"
fi
