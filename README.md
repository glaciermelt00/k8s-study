# Kubernetes 学習リポジトリ

このリポジトリでは、Kubernetes を体系的に学習するための 2 つのコースを提供しています。

## コース概要

### [📚 Kubernetes 入門コース](./k8s-intro/)

Kubernetes の基礎から実践的な運用まで、Minikube を使用してローカル環境で学習します。

**主な学習内容：**
- Pod、Deployment、Service などの基本リソース
- StatefulSet を使用したデータベース構築
- Ingress によるルーティング
- ConfigMap/Secret による設定管理
- Job/CronJob によるバッチ処理
- HPA による自動スケーリング

**対象者：**
- Kubernetes を初めて学ぶ方
- 基本的な概念を実践的に理解したい方

### [🚀 Kubernetes 応用コース](./k8s-advanced/)

AWS EKS 上で商用レベルの Kubernetes 環境を構築し、高度な運用テクニックを習得します。

**主な学習内容：**
- Kustomize によるマニフェスト管理
- Helm による OSS 導入
- Pod Identity による IAM 制御
- External Secrets Operator による機密情報管理
- Argo CD による GitOps
- Fluent Bit によるログ転送
- Prometheus/Grafana によるモニタリング
- Istio × ALB による高度なルーティング

**対象者：**
- 入門コースを修了した方
- 商用環境での Kubernetes 運用を学びたい方

## 学習の進め方

1. **入門コースから開始**
   - Kubernetes の基本概念を理解
   - ローカル環境で実践的な構築を体験

2. **応用コースへ進む**
   - AWS EKS を使用した本格的な環境構築
   - エンタープライズレベルのツールチェーンを習得

## リポジトリ構成

```
k8s-study/
├── README.md          # 本ファイル
├── k8s-intro/         # 入門コース
│   ├── deployments/   # Kubernetes マニフェスト
│   ├── apps/          # アプリケーションコード
│   ├── docs/          # ドキュメント
│   └── scripts/       # ユーティリティスクリプト
└── k8s-advanced/      # 応用コース
    ├── infrastructure/ # Terraform コード
    ├── kubernetes/     # K8s マニフェスト（Kustomize/Helm）
    ├── argocd/        # ArgoCD 設定
    └── docs/          # ドキュメント
```

## 必要な環境

### 入門コース
- Docker Desktop または Docker Engine
- Minikube
- kubectl

### 応用コース
- AWS アカウント
- AWS CLI
- eksctl
- helm
- kustomize
- terraform

## ライセンス

このプロジェクトは学習目的で公開されています。自由に参照・改変してご利用ください。

## 貢献

Issue や Pull Request は歓迎します。学習コンテンツの改善提案もお待ちしています。