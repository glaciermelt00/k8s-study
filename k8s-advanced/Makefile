.PHONY: help setup validate-env create-cluster delete-cluster deploy-stg deploy-prod build-image push-image

# デフォルトターゲット
help:
	@echo "Available commands:"
	@echo "  make setup              - 初期セットアップ"
	@echo "  make validate-env       - 環境変数の検証"
	@echo "  make create-cluster     - EKSクラスターの作成"
	@echo "  make delete-cluster     - EKSクラスターの削除"
	@echo "  make deploy-stg        - ステージング環境へのデプロイ"
	@echo "  make deploy-prod       - 本番環境へのデプロイ"
	@echo "  make build-image       - Dockerイメージのビルド"
	@echo "  make push-image        - ECRへのイメージプッシュ"

# 環境変数の検証
validate-env:
	@echo "Validating environment variables..."
	@test -n "$(AWS_ACCOUNT_ID)" || (echo "AWS_ACCOUNT_ID is not set" && exit 1)
	@test -n "$(AWS_REGION)" || (echo "AWS_REGION is not set" && exit 1)
	@test -n "$(EKS_CLUSTER_NAME)" || (echo "EKS_CLUSTER_NAME is not set" && exit 1)
	@echo "✓ All required environment variables are set"

# 初期セットアップ
setup: validate-env
	@echo "Setting up k8s-advanced environment..."
	@echo "Installing required tools..."
	@which eksctl > /dev/null || (echo "Please install eksctl" && exit 1)
	@which helm > /dev/null || (echo "Please install helm" && exit 1)
	@which kustomize > /dev/null || (echo "Please install kustomize" && exit 1)
	@echo "✓ All required tools are installed"

# EKSクラスターの作成
create-cluster: validate-env
	@echo "Creating EKS cluster $(EKS_CLUSTER_NAME)..."
	cd infrastructure/terraform/environments/stg && \
		terraform init && \
		terraform plan && \
		terraform apply -auto-approve
	aws eks update-kubeconfig --name $(EKS_CLUSTER_NAME) --region $(AWS_REGION)
	@echo "✓ EKS cluster created successfully"

# EKSクラスターの削除
delete-cluster: validate-env
	@echo "Deleting EKS cluster $(EKS_CLUSTER_NAME)..."
	cd infrastructure/terraform/environments/stg && \
		terraform destroy -auto-approve
	@echo "✓ EKS cluster deleted successfully"

# Kustomizeによるマニフェストのビルド確認
kustomize-build-stg:
	@echo "Building staging manifests..."
	kustomize build kubernetes/overlays/stg

kustomize-build-prod:
	@echo "Building production manifests..."
	kustomize build kubernetes/overlays/prod

# ステージング環境へのデプロイ
deploy-stg: validate-env
	@echo "Deploying to staging environment..."
	kubectl apply -k kubernetes/overlays/stg
	@echo "✓ Deployed to staging"

# 本番環境へのデプロイ
deploy-prod: validate-env
	@echo "Deploying to production environment..."
	@read -p "Are you sure you want to deploy to production? (yes/no): " confirm && \
		[ "$$confirm" = "yes" ] || (echo "Deployment cancelled" && exit 1)
	kubectl apply -k kubernetes/overlays/prod
	@echo "✓ Deployed to production"

# Dockerイメージのビルド
build-image:
	@echo "Building Docker image..."
	docker build -t $(ECR_REPOSITORY):latest ../k8s-intro/apps/api/
	@echo "✓ Docker image built"

# ECRへのイメージプッシュ
push-image: validate-env build-image
	@echo "Pushing image to ECR..."
	aws ecr get-login-password --region $(AWS_REGION) | \
		docker login --username AWS --password-stdin $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
	docker tag $(ECR_REPOSITORY):latest \
		$(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPOSITORY):latest
	docker push $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com/$(ECR_REPOSITORY):latest
	@echo "✓ Image pushed to ECR"

# Helmリポジトリの追加
helm-repos:
	@echo "Adding Helm repositories..."
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo add fluent https://fluent.github.io/helm-charts
	helm repo add aws https://aws.github.io/eks-charts
	helm repo update
	@echo "✓ Helm repositories added"

# 状態確認
status:
	@echo "=== Cluster Info ==="
	kubectl cluster-info
	@echo "\n=== Nodes ==="
	kubectl get nodes
	@echo "\n=== Namespaces ==="
	kubectl get namespaces
	@echo "\n=== Deployments ==="
	kubectl get deployments -A

# ログ確認
logs-stg:
	kubectl logs -n slack-metrics-stg -l app=slack-metrics-api --tail=100 -f

logs-prod:
	kubectl logs -n slack-metrics-prod -l app=slack-metrics-api --tail=100 -f

# クリーンアップ
clean:
	@echo "Cleaning up generated files..."
	find . -name "*.tmp" -delete
	@echo "✓ Cleanup complete"