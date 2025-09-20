# Kubernetes 運用監視ガイド

## 概要

このドキュメントでは、Kubernetes クラスターと Slack Metrics API の運用監視方法について説明します。

## 目次

1. [Metrics Server による リソースモニタリング](#metrics-server-による-リソースモニタリング)
2. [Health Check Probes の設定](#health-check-probes-の設定)
3. [リソース制限（Requests/Limits）](#リソース制限requestslimits)
4. [HPA（Horizontal Pod Autoscaler）](#hpahorizontal-pod-autoscaler)
5. [ログ収集と分析](#ログ収集と分析)
6. [監視コマンド集](#監視コマンド集)
7. [トラブルシューティング](#トラブルシューティング)

## Metrics Server による リソースモニタリング

### Metrics Server の有効化

```bash
# Metrics Server を有効化
minikube addons enable metrics-server

# 状態確認
kubectl get pods -n kube-system | grep metrics-server
```

### kubectl top コマンドの使用

```bash
# ノードのリソース使用状況
kubectl top nodes

# Pod のリソース使用状況（全 namespace）
kubectl top pods -A

# 特定の namespace の Pod
kubectl top pods -n default

# リソース使用量でソート
kubectl top pods --sort-by=cpu
kubectl top pods --sort-by=memory
```

## Health Check Probes の設定

### 実装済みの Probe

#### 1. Liveness Probe

Pod が正常に動作しているかを確認。失敗した場合は Pod を再起動。

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 30
  timeoutSeconds: 5
```

#### 2. Readiness Probe

Pod がトラフィックを受信する準備ができているかを確認。

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 10
  timeoutSeconds: 5
```

#### 3. Startup Probe

起動時間が長いアプリケーション用。起動が完了するまで他の Probe を無効化。

```yaml
startupProbe:
  exec:
    command:
      - pg_isready
      - -U
      - postgres
  initialDelaySeconds: 10
  periodSeconds: 10
  failureThreshold: 30
```

### Probe の状態確認

```bash
# Pod の詳細情報から Probe の状態を確認
kubectl describe pod <pod-name>

# Events セクションで Probe の失敗を確認
kubectl get events --field-selector involvedObject.name=<pod-name>
```

## リソース制限（Requests/Limits）

### 設定内容

| コンポーネント    | CPU Request | CPU Limit | Memory Request | Memory Limit |
| ----------------- | ----------- | --------- | -------------- | ------------ |
| slack-metrics-api | 250m        | 500m      | 64Mi           | 128Mi        |
| postgres          | 250m        | 500m      | 512Mi          | 1Gi          |
| cronjob           | 100m        | 200m      | 64Mi           | 128Mi        |

### リソース使用状況の監視

```bash
# 実際の使用量と制限値を比較
kubectl top pods

# リソース制限による throttling の確認
kubectl describe pod <pod-name> | grep -A 5 "Limits"
```

## HPA（Horizontal Pod Autoscaler）

### 設定内容

```yaml
対象: slack-metrics-api
最小レプリカ: 2
最大レプリカ: 10
CPU ターゲット: 50%
Memory ターゲット: 80%
```

### HPA の監視

```bash
# HPA の状態確認
kubectl get hpa slack-metrics-api-hpa

# 詳細情報の確認
kubectl describe hpa slack-metrics-api-hpa

# スケーリングイベントの確認
kubectl get events --field-selector involvedObject.name=slack-metrics-api-hpa
```

### 負荷テスト

```bash
# 負荷テストスクリプトの実行
chmod +x scripts/load-test.sh
./scripts/load-test.sh

# 環境変数でカスタマイズ
CONCURRENT_USERS=100 DURATION=600 ./scripts/load-test.sh
```

## ログ収集と分析

### 基本的なログ確認

```bash
# Pod のログ確認
kubectl logs <pod-name>

# 前のコンテナのログ（クラッシュした場合）
kubectl logs <pod-name> --previous

# ログのストリーミング
kubectl logs -f <pod-name>

# タイムスタンプ付きログ
kubectl logs <pod-name> --timestamps
```

### 複数 Pod のログ確認

```bash
# ラベルセレクタで複数 Pod のログを確認
kubectl logs -l app=slack-metrics-api --prefix=true

# 最新の 100 行のみ
kubectl logs -l app=slack-metrics-api --tail=100
```

### Stern を使用した高度なログ確認

Stern は複数 Pod のログを同時に確認できるツールです。

```bash
# Stern のインストール（macOS）
brew install stern

# 使用例
stern slack-metrics-api  # app名でフィルタ
stern -n default .      # namespace 内の全 Pod
stern --since 15m .     # 過去 15 分のログ
stern -t .              # タイムスタンプ付き
```

## 監視コマンド集

### リアルタイム監視

```bash
# Pod の状態を継続的に監視
watch -n 2 kubectl get pods

# リソース使用量を継続的に監視
watch -n 5 kubectl top pods

# HPA の状態を継続的に監視
watch -n 10 kubectl get hpa
```

### 診断コマンド

```bash
# クラスター全体の状態確認
kubectl cluster-info
kubectl get nodes
kubectl get all -A

# 問題のある Pod を特定
kubectl get pods -A | grep -v "Running\|Completed"

# Pod の詳細診断
kubectl describe pod <pod-name>
kubectl get pod <pod-name> -o yaml
```

### パフォーマンス分析

```bash
# CPU 使用率が高い Pod を特定
kubectl top pods --sort-by=cpu | head -10

# メモリ使用率が高い Pod を特定
kubectl top pods --sort-by=memory | head -10

# ノードのリソース配分を確認
kubectl describe nodes | grep -A 5 "Allocated resources"
```

## トラブルシューティング

### 1. Pod が起動しない

```bash
# Pod の状態確認
kubectl describe pod <pod-name>

# イベントログ確認
kubectl get events --sort-by=.lastTimestamp

# ログ確認（起動に失敗した場合）
kubectl logs <pod-name> --previous
```

### 2. リソース不足

```bash
# ノードのリソース状況確認
kubectl top nodes
kubectl describe nodes

# Pending Pod の理由確認
kubectl describe pod <pending-pod-name>
```

### 3. HPA が動作しない

```bash
# Metrics Server の状態確認
kubectl get pods -n kube-system | grep metrics-server

# HPA の詳細確認
kubectl describe hpa <hpa-name>

# メトリクスが取得できているか確認
kubectl get --raw /apis/metrics.k8s.io/v1beta1/pods
```

### 4. Probe の失敗

```bash
# Probe の履歴確認
kubectl describe pod <pod-name> | grep -A 10 "Liveness\|Readiness"

# エンドポイントの手動確認
kubectl exec <pod-name> -- curl localhost:8080/health
```

## ベストプラクティス

### 1. リソース設定

- **Requests**: 通常時の使用量に基づいて設定
- **Limits**: ピーク時の使用量 + 20% 程度の余裕を持たせる
- CPU Limits は慎重に設定（throttling の原因になる）

### 2. Probe の設定

- **initialDelaySeconds**: アプリケーションの起動時間を考慮
- **timeoutSeconds**: ネットワークレイテンシを考慮して設定
- **periodSeconds**: 頻繁すぎるチェックは負荷になるので注意

### 3. HPA の設定

- **stabilizationWindowSeconds**: 頻繁なスケーリングを防ぐ
- **scaleDown**: scaleUp より長い時間を設定してフラッピングを防ぐ
- 複数のメトリクスを組み合わせてより適切なスケーリングを実現

### 4. ログ管理

- 構造化ログ（JSON）の採用
- 適切なログレベルの設定
- センシティブな情報をログに含めない

## 高度な監視ツール

### k9s - Kubernetes TUI

```bash
# インストール（macOS）
brew install k9s

# 起動
k9s
```

主な操作:

- `:` でコマンドモード
- `/` で検索
- `l` でログ表示
- `d` で describe
- `ctrl+a` で全 namespace 表示

### kubectl プラグイン（krew）

```bash
# krew のインストール
brew install krew

# 便利なプラグイン
kubectl krew install neat    # YAML の整形
kubectl krew install tree    # リソースのツリー表示
kubectl krew install status  # 詳細なステータス表示
```

## まとめ

効果的な Kubernetes の運用監視には、以下の要素が重要です：

1. **プロアクティブな監視**: Metrics Server と kubectl top による定期的な確認
2. **適切な Health Check**: Liveness、Readiness、Startup Probe の使い分け
3. **リソース管理**: Requests/Limits の適切な設定
4. **自動スケーリング**: HPA による負荷に応じた自動調整
5. **ログ分析**: 問題の早期発見と原因究明

これらの機能を組み合わせることで、安定した Kubernetes クラスターの運用が可能になります。
