# プロジェクト構造（ローカル環境）

## ディレクトリ構成

```
k8s-study/
├── apps/                        # アプリケーションコード
│   ├── api/                     # API サーバー
│   │   ├── main.go
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── go.sum
│   │
│   └── migration/               # DB マイグレーション
│       ├── migrations/          # マイグレーションファイル
│       │   ├── 000001_init.up.sql
│       │   └── 000001_init.down.sql
│       ├── Dockerfile
│       ├── go.mod
│       └── go.sum
│
├── deployments/                 # Kubernetes マニフェスト
│   ├── postgres/                # PostgreSQL
│   │   ├── configmap.yaml
│   │   ├── service.yaml
│   │   ├── statefulset.yaml
│   │   └── client-pod.yaml
│   │
│   ├── api/                     # API サーバー
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   │
│   └── migration/               # マイグレーション Job
│       └── job.yaml
│
├── scripts/                     # ユーティリティスクリプト
│   └── create-secret.sh         # Secret 生成
│
├── docs/                        # ドキュメント
│   └── minikube-setup.md
│
├── .envrc.example               # 環境変数テンプレート
├── .envrc                       # 環境変数（.gitignore）
├── .gitignore
├── Makefile                     # タスクランナー
└── README.md
```

## 特徴

- **シンプルな構成**: ローカル開発に必要な要素のみ
- **明確な分離**: アプリケーション（apps）とインフラ（deployments）を分離
- **段階的な拡張**: 将来的に Kustomize や環境別設定を追加可能

## 次のステップ

1. DB マイグレーションツールの実装
2. API サーバーの基本実装
3. Makefile でのタスク自動化