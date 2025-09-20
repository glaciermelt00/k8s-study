# Kubernetes 入門コース 総括

このドキュメントは、Kubernetes 入門コースで学習した内容と実装したリソースをまとめたものです。

## 目次

1. [プロジェクト概要](#プロジェクト概要)
2. [実装した Kubernetes リソース](#実装したkubernetesリソース)
3. [Service の種類と実装例](#service-の種類と実装例)
4. [Ingress と LoadBalancer の違い](#ingress-と-loadbalancer-の違い)
5. [Kubernetes アーキテクチャの理解](#kubernetes-アーキテクチャの理解)
6. [学習成果と今後の展望](#学習成果と今後の展望)

## プロジェクト概要

本プロジェクトでは、Slack メトリクス収集・管理システムを題材に、実践的な Kubernetes の学習を行いました。

### システム構成

- **フロントエンド**: なし（API のみ）
- **バックエンド**: Node.js (Express) API サーバー
- **データベース**: PostgreSQL 15
- **バッチ処理**: Go 言語による定期処理
- **インフラ**: Minikube によるローカル Kubernetes クラスター

### 主な機能

1. Slack メトリクスの収集と保存
2. RESTful API によるデータアクセス
3. 定期的なメトリクス集計（CronJob）
4. 自動スケーリング（HPA）
5. ヘルスチェックとモニタリング

## 実装した Kubernetes リソース

### 1. ワークロードリソース

| リソースタイプ  | 名前                  | 用途                    | 特徴                        |
| --------------- | --------------------- | ----------------------- | --------------------------- |
| **StatefulSet** | postgres              | PostgreSQL データベース | 永続ストレージ、固定 Pod 名 |
| **Deployment**  | slack-metrics-api     | API サーバー            | ステートレス、レプリカ管理  |
| **Job**         | migration-job         | DB マイグレーション     | 一回限りの処理              |
| **CronJob**     | slack-metrics-cronjob | 定期バッチ処理          | スケジュール実行            |

### 2. ネットワークリソース

| リソースタイプ          | 名前                       | 用途                   |
| ----------------------- | -------------------------- | ---------------------- |
| **Service (Headless)**  | postgres-headless          | StatefulSet 用サービス |
| **Service (ClusterIP)** | slack-metrics-api-internal | 内部通信用             |
| **Service (NodePort)**  | slack-metrics-api          | 外部公開用             |
| **Ingress**             | slack-metrics-api-ingress  | HTTP ルーティング      |

### 3. 設定・ストレージリソース

| リソースタイプ | 名前                 | 用途               |
| -------------- | -------------------- | ------------------ |
| **ConfigMap**  | postgres-config      | PostgreSQL 設定    |
| **ConfigMap**  | slack-metrics-config | API 設定           |
| **Secret**     | postgres-secret      | DB 認証情報        |
| **PVC**        | postgres-pvc         | データベース永続化 |

### 4. スケーリング・監視

| リソースタイプ | 名前                  | 用途             |
| -------------- | --------------------- | ---------------- |
| **HPA**        | slack-metrics-api-hpa | 自動スケーリング |

## Service の種類と実装例

### 1. Headless Service

**実装例**: `postgres-headless`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: postgres-headless
spec:
  clusterIP: None # Headless Service の特徴
  selector:
    app: postgres
  ports:
    - port: 5432
      targetPort: 5432
```

**特徴**:

- ClusterIP が `None` に設定される
- DNS により各 Pod の IP アドレスが直接解決される
- StatefulSet と組み合わせて使用
- Pod 名での直接アクセスが可能（例: postgres-0.postgres-headless）

**使用場面**:

- ステートフルなアプリケーション
- Pod 間の直接通信が必要な場合
- マスター・スレーブ構成のデータベース

### 2. ClusterIP Service

**実装例**: `slack-metrics-api-internal`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: slack-metrics-api-internal
spec:
  type: ClusterIP # デフォルト値
  selector:
    app: slack-metrics-api
  ports:
    - port: 8080
      targetPort: 8080
```

**特徴**:

- クラスター内部からのみアクセス可能
- 仮想 IP アドレスが割り当てられる
- ロードバランシング機能を提供
- 最も一般的な Service タイプ

**使用場面**:

- マイクロサービス間の内部通信
- データベースへのアクセス
- 外部公開の必要がないサービス

### 3. NodePort Service

**実装例**: `slack-metrics-api`

```yaml
apiVersion: v1
kind: Service
metadata:
  name: slack-metrics-api
spec:
  type: NodePort
  selector:
    app: slack-metrics-api
  ports:
    - port: 8080
      targetPort: 8080
      nodePort: 32432 # 30000-32767 の範囲
```

**特徴**:

- 各ノードの特定ポートで外部からアクセス可能
- ClusterIP の機能も含む
- ポート範囲は 30000-32767
- 開発・テスト環境でよく使用

**使用場面**:

- 開発環境での外部アクセス
- Ingress Controller が使えない環境
- 簡易的な外部公開

### 4. LoadBalancer Service

**実装状況**: 本プロジェクトでは未実装（Minikube の制約）

```yaml
# 実装例（クラウド環境）
apiVersion: v1
kind: Service
metadata:
  name: example-loadbalancer
spec:
  type: LoadBalancer
  selector:
    app: example
  ports:
    - port: 80
      targetPort: 8080
```

**特徴**:

- クラウドプロバイダーのロードバランサーと統合
- 外部 IP アドレスが自動的に割り当てられる
- NodePort と ClusterIP の機能も含む
- 本番環境での主要な公開方法

**使用場面**:

- 本番環境での外部公開
- 高可用性が必要なサービス
- クラウド環境での運用

## Ingress と LoadBalancer の違い

### 実装した Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: slack-metrics-api-ingress
spec:
  ingressClassName: nginx
  rules:
    - host: sm-api.local
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: slack-metrics-api-internal
                port:
                  number: 8080
```

### 主な違い

| 項目             | Ingress                 | LoadBalancer Service  |
| ---------------- | ----------------------- | --------------------- |
| **OSI レイヤー** | L7 (アプリケーション層) | L4 (トランスポート層) |
| **プロトコル**   | HTTP/HTTPS              | TCP/UDP               |
| **ルーティング** | ホスト名、パスベース    | ポートベース          |
| **SSL/TLS 終端** | 可能                    | 別途設定が必要        |
| **コスト**       | 1 つで複数サービス対応  | サービスごとに必要    |
| **設定の柔軟性** | 高い                    | 低い                  |

### Ingress の利点

1. **効率的なリソース利用**

   - 1 つの IP アドレスで複数のサービスを公開
   - コスト効率が良い

2. **高度なルーティング**

   - ホスト名ベース: `api.example.com`, `web.example.com`
   - パスベース: `/api/*`, `/static/*`

3. **SSL/TLS の一元管理**

   - Ingress Controller で証明書を管理
   - 自動更新（cert-manager との連携）

4. **追加機能**
   - リクエストの書き換え
   - レート制限
   - 認証・認可

## Kubernetes アーキテクチャの理解

### Control Plane コンポーネント

1. **kube-apiserver**

   - Kubernetes API のフロントエンド
   - すべての操作の入り口
   - 認証・認可・アドミッションコントロール

2. **etcd**

   - 分散型 Key-Value ストア
   - クラスターの状態を保存
   - 高可用性のための複数レプリカ

3. **kube-scheduler**

   - Pod の配置先ノードを決定
   - リソース要求、アフィニティルールを考慮
   - カスタムスケジューラーも可能

4. **kube-controller-manager**
   - 各種コントローラーを実行
   - ReplicaSet、Deployment、Service などの管理
   - 望ましい状態の維持

### Worker Node コンポーネント

1. **kubelet**

   - 各ノードで実行されるエージェント
   - Pod の起動・監視
   - コンテナランタイムとの通信

2. **kube-proxy**

   - ネットワークプロキシ
   - Service の実装
   - iptables/IPVS によるロードバランシング

3. **Container Runtime**
   - コンテナの実行環境
   - Docker、containerd、CRI-O など
   - CRI (Container Runtime Interface) 準拠

### ネットワークアーキテクチャ

```
外部トラフィック
    ↓
[Ingress Controller]
    ↓
[Service (ClusterIP)]
    ↓
[Pod Network]
    ↓
[Container]
```

### ストレージアーキテクチャ

```
[Pod] → [PVC] → [PV] → [実際のストレージ]
         ↑
    [StorageClass]
    (動的プロビジョニング)
```

## 学習成果と今後の展望

### 習得したスキル

1. **基本的なリソース管理**

   - YAML マニフェストの作成
   - kubectl による操作
   - リソースの依存関係の理解

2. **ネットワーキング**

   - Service の種類と使い分け
   - Ingress による HTTP ルーティング
   - DNS とサービスディスカバリ

3. **ステート管理**

   - StatefulSet による永続化
   - ConfigMap/Secret の活用
   - データベースの運用

4. **運用・監視**
   - Health Check (Liveness/Readiness Probe)
   - リソース制限と HPA
   - ログ収集とトラブルシューティング

### 実装の工夫点

1. **環境変数の外部化**

   - `.envrc` による秘匿情報管理
   - ConfigMap による設定の分離

2. **自動化スクリプト**

   - Makefile によるタスク自動化
   - シークレット生成スクリプト

3. **モニタリング体制**
   - Metrics Server の活用
   - 包括的な監視ガイドの作成

### 今後の学習計画

1. **セキュリティ強化**

   - RBAC (Role-Based Access Control)
   - NetworkPolicy
   - Pod Security Standards

2. **高度な運用**

   - GitOps (ArgoCD/Flux)
   - Service Mesh (Istio/Linkerd)
   - Observability (Prometheus/Grafana)

3. **マルチクラスター**
   - クラスター間通信
   - フェデレーション
   - マルチリージョン展開

### まとめ

このプロジェクトを通じて、Kubernetes の基本的な概念から実践的な運用まで幅広く学習することができました。特に、Service の種類による使い分け、Ingress と LoadBalancer の違い、そして Kubernetes の内部アーキテクチャについて深い理解を得ることができました。

今後は、この基礎知識を土台として、より高度な Kubernetes の機能やエコシステムツールの学習を進めていく予定です。
