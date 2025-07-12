# Minikube で StatefulSet を動かす手順

## 前提条件

- Docker Desktop がインストールされていること
- `kubectl` がインストールされていること
- `minikube` がインストールされていること

## 1. Minikube の起動

```bash
# Minikube クラスタを起動（メモリとCPUを多めに割り当て）
minikube start --cpus=2 --memory=4096 --driver=docker

# クラスタの状態を確認
minikube status
```

## 2. ストレージの準備

StatefulSet では PersistentVolume が必要です。Minikube では自動的に `hostPath` タイプの PV が作成されます。

```bash
# デフォルトの StorageClass を確認
kubectl get storageclass
```

## 3. PostgreSQL StatefulSet のデプロイ

### 環境変数の設定

```bash
# プロジェクトルートに移動
cd /Volumes/dev/git-dev/k8s-study

# 環境変数を設定
cp .envrc.example .envrc
vim .envrc

# direnv を使用している場合
direnv allow
```

### Secret の生成

```bash
# Secret マニフェストを生成
./scripts/create-secret.sh
```

### リソースのデプロイ

```bash
# すべてのリソースを適用
kubectl apply -f postgres/

# デプロイの確認
kubectl get all -l app=postgres
```

## 4. 動作確認

### Pod の状態確認

```bash
# Pod の詳細情報
kubectl describe pod postgres-0

# ログの確認
kubectl logs postgres-0
```

### データベースへの接続

```bash
# Pod 内から接続
kubectl exec -it postgres-0 -- psql -U postgres -d postgresdb

# ポートフォワード経由で接続
kubectl port-forward postgres-0 15432:5432 &
psql -h localhost -U postgres -d postgresdb
```

### TablePlus からの接続

1. ポートフォワードを開始：

   ```bash
   # ローカルの 15432 ポートを Pod の 5432 ポートに転送
   kubectl port-forward postgres-0 15432:5432
   ```

2. TablePlus で新しい接続を作成：

   - **Host**: `localhost` または `127.0.0.1`
   - **Port**: `15432`
   - **User**: `postgres`（.envrc で設定した値）
   - **Password**: `postgres123`（.envrc で設定した値）
   - **Database**: `postgresdb`（.envrc で設定した値）

3. 「Test」ボタンで接続テスト後、「Connect」で接続

### コマンドラインからの接続確認

```bash
# ポートフォワード経由で接続
psql -h localhost -p 15432 -U postgres -d postgresdb
```

### Headless Service 経由でのアクセス

StatefulSet の Pod は Headless Service を通じて DNS 名でアクセスできます。

#### DNS 名の形式

```
<pod-name>.<service-name>.<namespace>.svc.cluster.local
```

PostgreSQL の場合：

- `postgres-0.postgres-headless.default.svc.cluster.local`

#### テスト用クライアント Pod の作成

```bash
# クライアント Pod を作成
kubectl apply -f postgres/test-pod.yaml

# クライアント Pod から Headless Service 経由で接続
kubectl exec -it postgres-client -- psql -h postgres-0.postgres-headless.default.svc.cluster.local -U postgres -d postgresdb

# または短縮形で接続
kubectl exec -it postgres-client -- psql -h postgres-0.postgres-headless -U postgres -d postgresdb

# Service 名だけでも接続可能（ラウンドロビンではなく、StatefulSet の場合は postgres-0 に接続）
kubectl exec -it postgres-client -- psql -h postgres-headless -U postgres -d postgresdb

# DNS 解決の確認
kubectl exec -it postgres-client -- nslookup postgres-headless
kubectl exec -it postgres-client -- nslookup postgres-0.postgres-headless
```

#### クリーンアップ

```bash
kubectl delete pod postgres-client
```

## 5. 永続性の確認

```bash
# データを作成
kubectl exec -it postgres-0 -- psql -U postgres -d postgresdb -c "CREATE TABLE test (id int);"
kubectl exec -it postgres-0 -- psql -U postgres -d postgresdb -c "INSERT INTO test VALUES (1);"

# Pod を削除
kubectl delete pod postgres-0

# Pod が再作成されるのを待つ
kubectl wait --for=condition=ready pod/postgres-0 --timeout=60s

# データが残っていることを確認
kubectl exec -it postgres-0 -- psql -U postgres -d postgresdb -c "SELECT * FROM test;"
```

## 6. トラブルシューティング

### PVC が Pending の場合

```bash
# PVC の状態を確認
kubectl get pvc

# イベントを確認
kubectl describe pvc postgres-data-postgres-0
```

### Pod が起動しない場合

```bash
# Pod のイベントを確認
kubectl describe pod postgres-0

# ログを確認
kubectl logs postgres-0 --previous
```

## 7. クリーンアップ

```bash
# リソースの削除
kubectl delete -f postgres/

# PVC の削除（データも削除される）
kubectl delete pvc -l app=postgres

# Minikube の停止
minikube stop

# Minikube クラスタの削除（完全にクリーンアップ）
minikube delete
```

## 注意事項

- Minikube はシングルノードクラスタのため、本番環境の挙動とは異なる場合があります
- `hostPath` ボリュームは Minikube VM 内に保存されるため、`minikube delete` でデータが失われます
- メモリ不足の場合は `minikube start` のメモリ割り当てを増やしてください
