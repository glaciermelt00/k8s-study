# Kubernetes アーキテクチャ図

このドキュメントでは、本プロジェクトで構築したシステムのアーキテクチャを視覚的に表現します。

## システム全体図

```mermaid
graph TB
    subgraph "外部"
        User[ユーザー]
        Browser[ブラウザ]
    end

    subgraph "Minikube Cluster"
        subgraph "Ingress Layer"
            Ingress[Ingress<br/>sm-api.local]
        end

        subgraph "Service Layer"
            SvcNodePort[Service NodePort<br/>slack-metrics-api<br/>:32432]
            SvcClusterIP[Service ClusterIP<br/>slack-metrics-api-internal<br/>:8080]
            SvcHeadless[Service Headless<br/>postgres-headless]
        end

        subgraph "Workload Layer"
            subgraph "API Pods"
                API1[Pod: API-1<br/>slack-metrics-api]
                API2[Pod: API-2<br/>slack-metrics-api]
            end

            subgraph "Database"
                PG[Pod: postgres-0<br/>PostgreSQL]
            end

            subgraph "Jobs"
                Migration[Job<br/>migration-job]
                CronJob[CronJob<br/>slack-metrics-cronjob]
            end
        end

        subgraph "Storage Layer"
            PVC[PVC<br/>postgres-data]
            PV[PV<br/>10Gi]
        end

        subgraph "Config Layer"
            CM1[ConfigMap<br/>postgres-config]
            CM2[ConfigMap<br/>slack-metrics-config]
            Secret[Secret<br/>postgres-secret]
        end

        HPA[HPA<br/>slack-metrics-api-hpa]
    end

    User --> Browser
    Browser --> |http://sm-api.local| Ingress
    Browser --> |:32432| SvcNodePort

    Ingress --> SvcClusterIP
    SvcNodePort --> API1
    SvcNodePort --> API2
    SvcClusterIP --> API1
    SvcClusterIP --> API2

    API1 --> SvcHeadless
    API2 --> SvcHeadless
    SvcHeadless --> PG

    Migration --> SvcHeadless
    CronJob --> SvcClusterIP

    PG --> PVC
    PVC --> PV

    API1 -.-> CM2
    API2 -.-> CM2
    API1 -.-> Secret
    API2 -.-> Secret
    PG -.-> CM1
    PG -.-> Secret

    HPA --> |スケーリング| API1
    HPA --> |スケーリング| API2
```

## ネットワークフロー図

```
┌─────────────────────────────────────────────────────────────────┐
│                          外部ネットワーク                          │
└─────────────────────┬───────────────────┬───────────────────────┘
                      │                   │
                      ▼                   ▼
            ┌─────────────────┐   ┌──────────────┐
            │ Ingress (L7)    │   │ NodePort     │
            │ sm-api.local:80 │   │ :32432       │
            └────────┬────────┘   └──────┬───────┘
                     │                   │
    ┌────────────────┴───────────────────┴────────────────┐
    │                  Service Layer (L4)                  │
    │  ┌─────────────────────────────────────────────┐    │
    │  │ slack-metrics-api-internal (ClusterIP)      │    │
    │  │ 10.96.31.4:8080                             │    │
    │  └─────────────────┬───────────────────────────┘    │
    └────────────────────┼────────────────────────────────┘
                         │
    ┌────────────────────┼────────────────────────────────┐
    │                Pod Network                          │
    │  ┌─────────────┴────────────┐                      │
    │  │    API Pods (2 replicas)  │                     │
    │  │    10.244.0.75-76:8080    │                     │
    │  └─────────────┬────────────┘                      │
    │                │                                    │
    │                ▼                                    │
    │  ┌──────────────────────────┐                      │
    │  │ postgres-headless Service │                      │
    │  └─────────────┬────────────┘                      │
    │                │                                    │
    │                ▼                                    │
    │  ┌──────────────────────────┐                      │
    │  │    PostgreSQL Pod         │                      │
    │  │    10.244.0.82:5432       │                      │
    │  └──────────────────────────┘                      │
    └─────────────────────────────────────────────────────┘
```

## Kubernetes コンポーネント配置図

```
┌─────────────────────── Minikube Node ───────────────────────┐
│                                                              │
│  Control Plane                    Worker Components          │
│  ┌────────────────┐               ┌────────────────┐        │
│  │ kube-apiserver │               │    kubelet     │        │
│  └────────┬───────┘               └────────┬───────┘        │
│           │                                 │                │
│  ┌────────▼───────┐               ┌────────▼───────┐        │
│  │      etcd      │               │   kube-proxy   │        │
│  └────────────────┘               └────────────────┘        │
│                                                              │
│  ┌────────────────┐               ┌────────────────┐        │
│  │ kube-scheduler │               │ Container      │        │
│  └────────────────┘               │ Runtime        │        │
│                                   │ (Docker)       │        │
│  ┌────────────────┐               └────────┬───────┘        │
│  │ kube-controller│                        │                │
│  │ -manager       │               ┌────────▼───────┐        │
│  └────────────────┘               │     Pods       │        │
│                                   └────────────────┘        │
│  ┌────────────────┐                                         │
│  │ CoreDNS        │                                         │
│  └────────────────┘                                         │
│                                                              │
│  ┌────────────────┐                                         │
│  │ Metrics Server │                                         │
│  └────────────────┘                                         │
└──────────────────────────────────────────────────────────────┘
```

## リソース依存関係図

```mermaid
graph LR
    subgraph "設定"
        Secret[Secret]
        CM[ConfigMap]
    end

    subgraph "ストレージ"
        PVC[PVC]
        PV[PV]
    end

    subgraph "ネットワーク"
        Ingress[Ingress]
        Service[Service]
    end

    subgraph "ワークロード"
        Deployment[Deployment]
        StatefulSet[StatefulSet]
        Job[Job]
        CronJob[CronJob]
        Pod[Pod]
    end

    subgraph "スケーリング"
        HPA[HPA]
    end

    Secret --> Pod
    CM --> Pod

    Deployment --> Pod
    StatefulSet --> Pod
    Job --> Pod
    CronJob --> Job

    Service --> Pod
    Ingress --> Service

    StatefulSet --> PVC
    PVC --> PV
    Pod --> PVC

    HPA --> Deployment

    style Pod fill:#f9f,stroke:#333,stroke-width:4px
```

## データフロー図

```
┌──────────┐     HTTP Request      ┌─────────┐
│  Client  │ ──────────────────▶  │ Ingress │
└──────────┘                       └────┬────┘
                                        │
                                        ▼
                              ┌─────────────────┐
                              │ Service (L4 LB) │
                              └────────┬────────┘
                                       │
                          ┌────────────┴────────────┐
                          ▼                        ▼
                  ┌───────────────┐        ┌───────────────┐
                  │   API Pod 1   │        │   API Pod 2   │
                  └───────┬───────┘        └───────┬───────┘
                          │                        │
                          └────────┬───────────────┘
                                   │
                                   ▼ SQL Query
                          ┌─────────────────┐
                          │ PostgreSQL Pod  │
                          └────────┬────────┘
                                   │
                                   ▼ Write
                          ┌─────────────────┐
                          │ Persistent      │
                          │ Volume          │
                          └─────────────────┘

データの流れ:
1. クライアントからの HTTP リクエスト
2. Ingress でホスト名/パスベースのルーティング
3. Service でロードバランシング
4. API Pod でビジネスロジック処理
5. PostgreSQL へのデータ永続化
6. レスポンスを逆順で返却
```

## CronJob 実行フロー

```
┌─────────────┐     毎日 9:00 (JST)    ┌──────────────┐
│ CronJob     │ ────────────────────▶ │ Job 作成      │
└─────────────┘                       └──────┬───────┘
                                              │
                                              ▼
                                      ┌───────────────┐
                                      │ Pod 起動      │
                                      └──────┬────────┘
                                              │
                                              ▼
                                      ┌───────────────┐
                                      │ バッチ処理実行 │
                                      └──────┬────────┘
                                              │
                          ┌───────────────────┴───────────────────┐
                          ▼                                       ▼
                  ┌───────────────┐                       ┌──────────────┐
                  │ API 呼び出し   │                       │ Slack 通知    │
                  │ (Internal Svc) │                       │ (外部 API)    │
                  └───────┬───────┘                       └──────────────┘
                          │
                          ▼
                  ┌───────────────┐
                  │ DB 更新        │
                  │ (postgres)     │
                  └───────────────┘
```

## セキュリティ境界

```
┌─────────────────────────────────────────────────────────┐
│                     インターネット                        │
└────────────────────────┬────────────────────────────────┘
                         │ ファイアウォール
                         ▼
┌─────────────────────────────────────────────────────────┐
│                  Ingress Controller                      │
│                  - SSL/TLS 終端                          │
│                  - ホスト名検証                          │
└────────────────────────┬────────────────────────────────┘
                         │
┌────────────────────────┼────────────────────────────────┐
│  Kubernetes Cluster    │                                │
│  ┌─────────────────────▼──────────────────────┐        │
│  │ Service Network (内部通信のみ)              │        │
│  │ - ClusterIP による分離                      │        │
│  │ - NetworkPolicy (未実装)                    │        │
│  └─────────────────────┬──────────────────────┘        │
│                        │                                │
│  ┌─────────────────────▼──────────────────────┐        │
│  │ Pod Security                                │        │
│  │ - SecurityContext                           │        │
│  │ - 非 root ユーザー実行                      │        │
│  │ - Read-only ファイルシステム                 │        │
│  └────────────────────────────────────────────┘        │
│                                                         │
│  ┌────────────────────────────────────────────┐        │
│  │ Secret Management                           │        │
│  │ - 環境変数による注入                         │        │
│  │ - Base64 エンコード（暗号化ではない）        │        │
│  └────────────────────────────────────────────┘        │
└─────────────────────────────────────────────────────────┘
```

## まとめ

このアーキテクチャ図は、Kubernetes 入門コースで構築したシステムの全体像を示しています。主要なポイント：

1. **多層構造**: Ingress → Service → Pod → Storage の階層的な構成
2. **高可用性**: HPA による自動スケーリング、複数レプリカの配置
3. **永続性**: StatefulSet と PVC による データベースの永続化
4. **定期処理**: CronJob によるバッチ処理の自動化
5. **設定管理**: ConfigMap と Secret による設定の外部化

このアーキテクチャは、実際のプロダクション環境でも応用可能な基本的なパターンを網羅しています。
