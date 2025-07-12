.PHONY: help setup start stop clean deploy-postgres deploy-api deploy-migration port-forward logs

# デフォルトターゲット
help:
	@echo "使用可能なコマンド:"
	@echo "  make setup              - 初期セットアップ（環境変数、Secret作成）"
	@echo "  make start              - Minikube起動"
	@echo "  make stop               - Minikube停止"
	@echo "  make clean              - クリーンアップ（リソース削除）"
	@echo "  make deploy-postgres    - PostgreSQLのデプロイ"
	@echo "  make deploy-api         - APIサーバーのデプロイ"
	@echo "  make deploy-migration   - マイグレーション実行"
	@echo "  make port-forward       - PostgreSQLポートフォワード"
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

# マイグレーション実行（実装後に使用）
deploy-migration:
	@echo "マイグレーションを実行します..."
	# kubectl apply -f deployments/migration/

# ポートフォワード
port-forward:
	@echo "PostgreSQL (15432:5432) のポートフォワードを開始します..."
	kubectl port-forward postgres-0 15432:5432

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