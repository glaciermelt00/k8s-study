# Kubernetes 応用コース

## 概要

Kubernetes 応用コースでは、入門編で学んだ Kubernetes の基礎を土台に、より実践的かつ商用レベルの運用に耐える構成を構築していきます。AWS の EKS クラスター上で、高度な構成や運用ノウハウを段階的に学習します。

## コース内容

### 1. Kustomize によるマニフェスト管理

共通の設定を `base/` に、環境固有の設定を `overlays/` に分離。patches や replacements を活用し、DRY な構成を実現します。

### 2. Helm による OSS 導入

Prometheus、Grafana、Fluent Bit などのモニタリングツールを Helm チャートで効率よく導入。`values.yaml` による環境ごとのカスタマイズにも対応します。

### 3. Pod Identity による IAM 制御

アプリケーションコードに認証処理を書くことなく、Pod 単位で IAM ロールを割り当て、AWS リソースへのセキュアなアクセスを実現します。

### 4. External Secrets Operator による機密情報管理

AWS Secrets Manager から Kubernetes の Secret へ自動同期。環境をまたいだシークレットの一元管理を行い、安全かつスケーラブルな設計を習得します。

### 5. Argo CD Image Updater による CI/CD 自動化

ECR への push をトリガーに、Argo CD Image Updater が GitHub に新しいタグを自動コミットし、Argo CD がその差分を即時適用。完全自動の CD パイプラインを構築します。

### 6. ログの転送

Fluent Bit で各ノードのログを収集、加工、フィルタリングし、CloudWatch Logs へ転送します。多くのエンジニアが苦手とする Fluent Bit の conf ファイルについても手を動かす課題を通してマスターしていきます。

### 7. メトリクスの収集と可視化

Prometheus と Grafana を用いて、Kubernetes 上で動作するアプリケーションや各種リソースのメトリクスを収集・可視化していきます。

具体的には、Prometheus を StatefulSet として自前でデプロイし、Service Discovery 機能を活用して各 Pod の CPU 使用率・メモリ使用量・HTTP レイテンシーなどの指標を取得します。

収集したデータは Grafana に連携し、実際に PromQL を使ってダッシュボードを作成し、リアルタイムでの可観測性を高めていきます。

### 8. Istio × ALB Controller による L7 ルーティング

Istio と ALB を連携させ、L7 レベルのきめ細かなルーティング制御を実現していきます。

AWS の ALB は、TLS 終端を行うエントリーポイントとして機能し、ALB Controller が Kubernetes リソース（Istio の Ingress や Service など）をもとにリスナールールやターゲットグループの設定を自動生成します。

ALB からのリクエストは、Istio Ingress Gateway の Pod に転送され、その後 Istio の Gateway リソースでドメイン・ポート・プロトコルを定義し、VirtualService でパスやヘッダーに応じたルーティングポリシーを構成します。

この構成により、以下のような高度なルーティングが可能になります。

- パスベース / ヘッダーベースのルーティング
- 特定ユーザーや環境にのみ特定バージョンを配信するカナリアリリース
- マイクロサービス間通信の可視化とセキュリティ制御（mTLS）

ALB のマネージド性と Istio の柔軟性を組み合わせることで、クラウドネイティブな商用レベルのトラフィックコントロールを Kubernetes 環境上で実現できるチャプターとなっています。

### 9. 総集編：本番環境の構築

Kubernetes 応用コースの総集編として本番環境の構築に挑戦します。

これまで構築してきた stg 環境の Terraform のコードや Kubernetes のマニフェストファイルをベースに複製していきます。

私自身も環境をまるっと複製する経験を通して、Kubernetes 力をつけることができました。このチャプターを通して、これまで断片的に身についてきた Kubernetes の知識が一気に自分の骨肉として武器になるはずです！

## ディレクトリ構成

```
k8s-advanced/
├── README.md                    # 本ファイル
├── infrastructure/              # インフラ構成
│   └── terraform/
│       ├── environments/        # 環境別設定
│       │   ├── stg/            # ステージング環境
│       │   └── prod/           # 本番環境
│       └── modules/            # 共通モジュール
├── kubernetes/                  # K8s マニフェスト
│   ├── base/                   # Kustomize base
│   │   ├── api/                # API アプリケーション
│   │   ├── postgres/           # PostgreSQL
│   │   └── monitoring/         # モニタリング設定
│   ├── overlays/               # 環境別設定
│   │   ├── stg/               # ステージング用
│   │   └── prod/              # 本番用
│   └── helm/                   # Helm チャート
│       ├── prometheus/         # Prometheus 設定
│       ├── grafana/           # Grafana 設定
│       └── fluent-bit/        # Fluent Bit 設定
├── argocd/                     # ArgoCD 設定
│   ├── applications/           # Application 定義
│   └── image-updater/         # Image Updater 設定
└── docs/                       # ドキュメント
    ├── 01-kustomize.md         # Kustomize ガイド
    ├── 02-helm.md             # Helm ガイド
    ├── 03-pod-identity.md     # Pod Identity ガイド
    ├── 04-external-secrets.md # External Secrets ガイド
    ├── 05-argocd.md          # ArgoCD ガイド
    ├── 06-logging.md         # ログ転送ガイド
    ├── 07-monitoring.md      # モニタリングガイド
    ├── 08-istio-alb.md      # Istio × ALB ガイド
    └── 09-production.md      # 本番環境構築ガイド
```

## 前提条件

### 必要なツール

- AWS CLI v2
- kubectl 1.28+
- eksctl
- helm 3.x
- kustomize 5.x
- terraform 1.5+
- direnv

### AWS アカウント

- EKS、ECR、Secrets Manager、CloudWatch などの権限が必要です
- 詳細は各チャプターのドキュメントを参照してください

## クイックスタート

1. **環境変数の設定**

   ```bash
   cp .envrc.example .envrc
   vim .envrc  # AWS認証情報などを設定
   direnv allow
   ```

2. **EKS クラスターの作成**

   ```bash
   cd infrastructure/terraform/environments/stg
   terraform init
   terraform plan
   terraform apply
   ```

3. **kubectl の設定**

   ```bash
   aws eks update-kubeconfig --name k8s-advanced-stg --region ap-northeast-1
   ```

4. **最初のチャプターを開始**

   ```bash
   cd docs
   cat 01-kustomize.md
   ```

## 学習の進め方

1. 各チャプターのドキュメントを読む
2. 実際にコマンドを実行して動作を確認
3. 課題に取り組む
4. 次のチャプターへ進む

順番通りに進めることを推奨しますが、特定のトピックに興味がある場合は該当チャプターから始めても構いません。

## サポート

質問や問題がある場合は、Issue を作成してください。

## ライセンス

このプロジェクトは学習目的で公開されています。