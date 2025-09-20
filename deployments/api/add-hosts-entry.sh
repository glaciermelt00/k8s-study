#!/bin/bash

# /etc/hosts にエントリを追加するスクリプト
# 管理者権限で実行してください

echo "Adding sm-api.local to /etc/hosts..."

# 既存のエントリをチェック
if grep -q "sm-api.local" /etc/hosts; then
    echo "sm-api.local already exists in /etc/hosts"
else
    # minikube の IP アドレスを取得
    MINIKUBE_IP=$(minikube ip)
    echo "minikube IP: $MINIKUBE_IP"

    # hosts ファイルに追加
    echo "$MINIKUBE_IP  sm-api.local" | sudo tee -a /etc/hosts
    echo "Added: $MINIKUBE_IP  sm-api.local"
fi

# 確認
echo -e "\nCurrent sm-api.local entry:"
grep sm-api.local /etc/hosts
