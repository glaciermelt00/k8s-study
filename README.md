# k8s-study

Kubernetes 学習用リポジトリ

## 構築する環境の全体像

### 1. **インフラ構成**

- **Minikube**を使用したローカル Kubernetes 環境
- **Docker**コンテナランタイム
- MacBook 上で動作（192.168.49.2:80 でアクセス可能）

### 2. **アプリケーション構成**

- **Node.js**ベースの Web アプリケーション
- **/etc/hosts**による名前解決（http://sm-api.local）
- **Ingress**によるルーティング制御

### 3. **Kubernetes リソース構成**

#### **ネットワーク層**

- **Ingress**: 外部からの HTTP アクセスを管理
- **Service (NodePort)**: クラスター外部からのアクセスを可能に
- **Service (ClusterIP)**: クラスター内部での通信
- **Service (Headless)**: StatefulSet 用のサービス

#### **ワークロード層**

- **Deployment**: ステートレスなアプリケーションのデプロイ
- **StatefulSet**: ステートフルなアプリケーション（データベース等）
- **Job**: バッチ処理の実行
- **CronJob**: 定期的なタスクの実行
- slack-metrics API: アプリケーション内の API エンドポイント
- DB マイグレーション: アプリケーション内のマイグレーション機能

#### **ストレージ層**

- **Pod**: 基本的なコンテナ実行単位（複数種類）
- **PVC (Persistent Volume Claim)**: 永続ストレージの要求
- **PV (Persistent Volume)**: 実際の永続ストレージ
- postgres DB: アプリケーションが接続するデータベース

#### **設定管理**

- **ConfigMap**: 環境変数や設定ファイルの管理
- **Secret**: 機密情報の管理

## 現在の実装状況

### PostgreSQL StatefulSet

本番環境での External Secrets Operator の使用を見据えた、ローカル開発環境での PostgreSQL 構築。

#### ディレクトリ構成

```
postgres/
├── configmap.yaml      # PostgreSQL設定ファイル
├── service.yaml        # Headless Service（StatefulSet用）
├── statefulset.yaml    # PostgreSQL StatefulSet定義
├── create-secret.sh    # Secret生成スクリプト
└── secret.yaml         # 生成されるSecretマニフェスト（.gitignore対象）
```

#### セットアップ手順

1. **環境変数の設定**

   ```bash
   # .envrcファイルをコピーして編集
   cp .envrc.example .envrc
   vim .envrc

   # direnvを使用している場合
   direnv allow
   ```

2. **Secret の生成**

   ```bash
   # 環境変数からSecretマニフェストを生成
   ./postgres/create-secret.sh
   ```

3. **リソースのデプロイ**

   ```bash
   # すべてのリソースを適用
   kubectl apply -f postgres/

   # 確認
   kubectl get statefulset,svc,secret,pvc
   kubectl get pods -l app=postgres
   ```

#### 使用方法

- **Pod 内から接続**

  ```bash
  kubectl exec -it postgres-0 -- psql -U postgres -d postgresdb
  ```

- **ポートフォワード経由で接続**
  ```bash
  kubectl port-forward postgres-0 5432:5432
  psql -h localhost -U postgres -d postgresdb
  ```

#### セキュリティ考慮事項

- `.envrc` - 実際の認証情報（.gitignore で除外）
- `secret.yaml` - 生成される Secret ファイル（.gitignore で除外）
- 本番環境では、AWS Secrets Manager と External Secrets Operator を使用予定

#### 含まれるリソース

- **Secret**: DB 認証情報（環境変数から生成）
- **ConfigMap**: PostgreSQL 設定ファイル
- **Headless Service**: StatefulSet 用の Pod 間通信
- **StatefulSet**: PostgreSQL 15 Alpine 版、1 レプリカ
- **PersistentVolumeClaim**: 10Gi のデータ永続化

## ドキュメント

- [Minikube で StatefulSet を動かす手順](docs/minikube-setup.md)
