.PHONY: help setup start stop clean deploy-postgres deploy-api run-migration port-forward dashboard logs

# デフォルトターゲット
help:
	@echo "使用可能なコマンド:"
	@echo "  make setup              - 初期セットアップ（環境変数、Secret作成）"
	@echo "  make start              - Minikube起動"
	@echo "  make stop               - Minikube停止"
	@echo "  make clean              - クリーンアップ（リソース削除）"
	@echo "  make deploy-postgres    - PostgreSQLのデプロイ"
	@echo "  make deploy-api         - APIサーバーのデプロイ"
	@echo "  make run-migration      - マイグレーション実行"
	@echo "  make port-forward       - PostgreSQLポートフォワード"
	@echo "  make dashboard          - Kubernetesダッシュボード起動"
	@echo "  make logs               - ログ表示"

# Minikube管理
start:
	minikube start --cpus=2 --memory=4096 --driver=docker
	@echo "Minikubeが起動しました"

stop:
	minikube stop
	@echo "Minikubeを停止しました"

# セットアップ
setup: start
	@if [ ! -f .envrc ]; then \
		echo ".envrcファイルをコピーしています..."; \
		cp .envrc.example .envrc; \
		echo ".envrcを編集してください"; \
		exit 1; \
	fi
	@echo "Secretを生成しています..."
	./scripts/create-secret.sh
	@echo "セットアップが完了しました"

# PostgreSQLデプロイ
deploy-postgres:
	kubectl apply -f deployments/postgres/
	@echo "PostgreSQLのデプロイを待っています..."
	kubectl wait --for=condition=ready pod/postgres-0 --timeout=60s
	@echo "PostgreSQLが起動しました"

# APIデプロイ（実装後に使用）
deploy-api:
	@echo "APIサーバーをデプロイします..."
	# kubectl apply -f deployments/api/

# Migration API のビルドとデプロイ
build-migration:
	cd apps/migration && go mod download
	docker build -t migration-api:latest ./apps/migration
	minikube image load migration-api:latest

run-migration: build-migration
	kubectl delete job migration-job --ignore-not-found=true
	kubectl apply -f deployments/migration/job.yaml


# ポートフォワード
port-forward:
	@echo "PostgreSQL (15432:5432) のポートフォワードを開始します..."
	kubectl port-forward postgres-0 15432:5432

# ダッシュボード
dashboard:
	@echo "Kubernetes ダッシュボードを起動します..."
	minikube dashboard

# ログ表示
logs:
	kubectl logs -f postgres-0

# PostgreSQLログのみ
logs-postgres:
	kubectl logs -f postgres-0

# クリーンアップ
clean:
	kubectl delete -f deployments/postgres/ || true
	kubectl delete pvc -l app=postgres || true
	@echo "リソースを削除しました"

# 完全クリーンアップ
clean-all: clean
	minikube delete
	@echo "Minikubeクラスタを削除しました"
