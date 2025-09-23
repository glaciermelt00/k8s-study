# Chapter 01: Kustomize によるマニフェスト管理

## 概要

Kustomize は Kubernetes のマニフェストファイルをテンプレート化せずに、宣言的に管理・カスタマイズするためのツールです。本章では、base と overlays を使用した DRY な構成管理を学習します。

## 学習目標

- Kustomize の基本概念（base、overlays、patches）を理解する
- 環境ごとの設定を効率的に管理する方法を習得する
- patches と replacements を使った高度なカスタマイズを実践する

## Kustomize とは

Kustomize は以下の特徴を持つ Kubernetes ネイティブな設定管理ツールです：

1. **テンプレートフリー**: YAML に変数を埋め込まない
2. **宣言的**: すべての変更が明示的に記述される
3. **レイヤード**: base の上に overlays を重ねる構造
4. **kubectl 統合**: kubectl に組み込まれている

## ディレクトリ構成

```
kubernetes/
├── base/                    # 共通設定
│   └── api/
│       ├── deployment.yaml
│       ├── service.yaml
│       ├── serviceaccount.yaml
│       ├── configmap.yaml
│       └── kustomization.yaml
└── overlays/               # 環境別設定
    ├── stg/
    │   ├── kustomization.yaml
    │   ├── deployment-patch.yaml
    │   └── account-config.yaml
    └── prod/
        ├── kustomization.yaml
        └── deployment-patch.yaml
```

## 実践：base の作成

### 1. Deployment の基本定義

`kubernetes/base/api/deployment.yaml` を確認してください：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slack-metrics-api
spec:
  replicas: 2 # デフォルトのレプリカ数
  template:
    spec:
      containers:
        - name: api
          image: slack-metrics-api:latest # 基本イメージ
```

ポイント：

- 環境に依存しない共通設定を定義
- セキュリティ設定（SecurityContext）も base に含める
- リソース制限はデフォルト値を設定

### 2. kustomization.yaml の作成

`kubernetes/base/api/kustomization.yaml` では、このディレクトリのリソースを定義：

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

commonLabels:
  app: slack-metrics-api
  component: api
  managed-by: kustomize

resources:
  - deployment.yaml
  - service.yaml
  - serviceaccount.yaml
  - configmap.yaml
```

## 実践：overlays による環境別設定

### 1. ステージング環境の設定

`kubernetes/overlays/stg/kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

namespace: slack-metrics-stg

resources:
  - ../../base/api
  - account-config.yaml

patchesStrategicMerge:
  - deployment-patch.yaml
```

### 2. deployment-patch.yaml でのカスタマイズ

環境固有の設定を patch として定義：

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slack-metrics-api
spec:
  replicas: 1 # ステージングは1レプリカ
  template:
    spec:
      containers:
        - name: api
          image: 123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/slack-metrics-api:latest
          env:
            - name: DB_HOST
              value: postgres-stg.abc123.ap-northeast-1.rds.amazonaws.com
```

## 高度な機能：replacements

### ServiceAccount の IAM ロール ARN を動的に設定

```yaml
replacements:
  - source:
      kind: ConfigMap
      name: account-config
      fieldPath: data.aws_account_id
    targets:
      - select:
          kind: ServiceAccount
          name: slack-metrics-api
        fieldPaths:
          - metadata.annotations.[eks.amazonaws.com/role-arn]
        options:
          delimiter: ":"
          index: 4
```

この設定により、AWS アカウント ID を一元管理できます。

## コマンド実行

### 1. マニフェストのビルド

```bash
# ステージング環境のマニフェストを確認
kustomize build kubernetes/overlays/stg

# 本番環境のマニフェストを確認
kustomize build kubernetes/overlays/prod
```

### 2. 直接適用

```bash
# kubectl と統合されているので直接適用可能
kubectl apply -k kubernetes/overlays/stg
```

### 3. Dry-run で確認

```bash
kubectl apply -k kubernetes/overlays/stg --dry-run=client -o yaml
```

## ベストプラクティス

### 1. DRY 原則の徹底

- 共通設定は必ず base に配置
- 環境固有の値のみ overlays で上書き
- 重複を見つけたら base に移動を検討

### 2. patches の使い分け

- **patchesStrategicMerge**: 部分的な更新（推奨）
- **patchesJson6902**: JSON Patch 形式での詳細な制御
- **patches**: 汎用的なパッチ（新しい方式）

### 3. ConfigMapGenerator の活用

```yaml
configMapGenerator:
  - name: app-config
    behavior: merge
    literals:
      - LOG_LEVEL=debug
    files:
      - application.yaml
```

### 4. 命名規則

- base: 環境に依存しない一般的な名前
- overlays: 環境名を含む具体的な名前

## トラブルシューティング

### 1. パッチが適用されない

```bash
# パッチの適用順序を確認
kustomize build kubernetes/overlays/stg --enable-alpha-plugins | grep -A10 "name: slack-metrics-api"
```

### 2. リソースが見つからない

```bash
# 相対パスが正しいか確認
cd kubernetes/overlays/stg
ls -la ../../base/api/
```

### 3. replacement が動作しない

```bash
# source と target のフィールドパスを確認
kustomize build kubernetes/overlays/stg -o yaml | grep -E "(role-arn|aws_account_id)"
```

## 演習問題

### 演習 1: 本番環境の設定作成

`kubernetes/overlays/prod` ディレクトリに本番環境用の設定を作成してください：

要件：

- レプリカ数: 3
- リソース制限: CPU 1000m、メモリ 1Gi
- 環境変数 LOG_LEVEL: warning

### 演習 2: NetworkPolicy の追加

base に NetworkPolicy を追加し、ステージングではすべての通信を許可、本番では特定の namespace からのみ許可する設定を作成してください。

### 演習 3: HPA の環境別設定

HorizontalPodAutoscaler を base に追加し、環境ごとに異なるスケーリング基準を設定してください。

## まとめ

本章では Kustomize を使った宣言的な設定管理を学習しました：

- ✅ base と overlays によるレイヤード構成
- ✅ patches による柔軟なカスタマイズ
- ✅ replacements による動的な値の注入
- ✅ ConfigMapGenerator によるコンフィグ管理

次章では、Helm を使った OSS の導入方法を学習します。

## 参考資料

- [Kustomize 公式ドキュメント](https://kustomize.io/)
- [Kubernetes SIG-CLI Kustomize](https://github.com/kubernetes-sigs/kustomize)
- [Kustomize Best Practices](https://kubectl.docs.kubernetes.io/references/kustomize/)
- [📝 演習問題の解答例](./01-kustomize-practice.md)
