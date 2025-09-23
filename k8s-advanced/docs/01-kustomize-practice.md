# Kustomize 実践ガイド

## 演習問題の解答例

### 演習1: 本番環境の設定作成 ✅

以下のファイルを作成しました：

#### kubernetes/overlays/prod/kustomization.yaml
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: slack-metrics-prod
commonLabels:
  environment: production
resources:
  - ../../base/api
  - account-config.yaml
patchesStrategicMerge:
  - deployment-patch.yaml
  - networkpolicy-patch.yaml
  - hpa-patch.yaml
configMapGenerator:
  - name: slack-metrics-api-config
    behavior: merge
    literals:
      - LOG_LEVEL=warning
      - METRICS_TYPE=database
      - REPORT_FORMAT=detailed
      - ENVIRONMENT=production
```

#### kubernetes/overlays/prod/deployment-patch.yaml
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: slack-metrics-api
spec:
  replicas: 3
  template:
    spec:
      containers:
        - name: api
          image: 123456789012.dkr.ecr.ap-northeast-1.amazonaws.com/slack-metrics-api:v1.0.0
          resources:
            requests:
              cpu: 500m
              memory: 512Mi
            limits:
              cpu: 1000m
              memory: 1Gi
```

### 演習2: NetworkPolicy の追加 ✅

#### kubernetes/base/api/networkpolicy.yaml
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slack-metrics-api-netpol
spec:
  podSelector:
    matchLabels:
      app: slack-metrics-api
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - ports:
        - protocol: TCP
          port: 8080
  egress:
    # DNS解決
    - to:
        - namespaceSelector:
            matchLabels:
              name: kube-system
      ports:
        - protocol: TCP
          port: 53
        - protocol: UDP
          port: 53
    # データベース接続
    - ports:
        - protocol: TCP
          port: 5432
    # HTTPS（外部API）
    - ports:
        - protocol: TCP
          port: 443
```

#### ステージング環境（すべての通信を許可）
```yaml
# kubernetes/overlays/stg/networkpolicy-patch.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slack-metrics-api-netpol
spec:
  ingress:
    - {}
  egress:
    - {}
```

#### 本番環境（特定namespaceのみ許可）
```yaml
# kubernetes/overlays/prod/networkpolicy-patch.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: slack-metrics-api-netpol
spec:
  ingress:
    - from:
        - namespaceSelector:
            matchLabels:
              name: slack-metrics-prod
        - namespaceSelector:
            matchLabels:
              name: istio-system
      ports:
        - protocol: TCP
          port: 8080
```

### 演習3: HPA の環境別設定 ✅

#### kubernetes/base/api/hpa.yaml
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: slack-metrics-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: slack-metrics-api
  minReplicas: 2
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

#### ステージング環境
```yaml
# kubernetes/overlays/stg/hpa-patch.yaml
spec:
  minReplicas: 1
  maxReplicas: 3
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 80
```

#### 本番環境
```yaml
# kubernetes/overlays/prod/hpa-patch.yaml
spec:
  minReplicas: 3
  maxReplicas: 10
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 50
```

## Minikube での動作確認

### 1. namespace の作成
```bash
kubectl create namespace slack-metrics-stg
kubectl create namespace slack-metrics-prod
```

### 2. シークレットの作成
```bash
# 仮のシークレットを作成
kubectl create secret generic postgres-secret \
  --namespace=slack-metrics-stg \
  --from-literal=POSTGRES_USER=postgres \
  --from-literal=POSTGRES_PASSWORD=password \
  --from-literal=POSTGRES_DB=slack_metrics

kubectl create secret generic slack-secret \
  --namespace=slack-metrics-stg \
  --from-literal=webhook-url=https://hooks.slack.com/services/dummy

kubectl create secret generic slack-metrics-api-secret \
  --namespace=slack-metrics-stg \
  --from-literal=DB_USER=postgres \
  --from-literal=DB_PASSWORD=password \
  --from-literal=DB_NAME=slack_metrics
```

### 3. デプロイ実行
```bash
# ビルド確認
kubectl kustomize kubernetes/overlays/stg

# ドライラン
kubectl apply -k kubernetes/overlays/stg --dry-run=client

# 実際にデプロイ
kubectl apply -k kubernetes/overlays/stg

# 確認
kubectl get all,networkpolicy,hpa -n slack-metrics-stg
```

### 4. 環境間の違いを確認
```bash
# ステージング環境と本番環境の差分確認
diff <(kubectl kustomize kubernetes/overlays/stg) \
     <(kubectl kustomize kubernetes/overlays/prod)
```

## まとめ

本実習では以下を達成しました：

✅ **演習1**: 本番環境の完全な設定（レプリカ数3、リソース制限、ログレベル）
✅ **演習2**: NetworkPolicyの環境別制御（stg:全許可、prod:制限付き）
✅ **演習3**: HPAの環境別スケーリング設定

Kustomizeによって、環境ごとの差分を明確に管理しながら、DRYな構成を実現できることが確認できました。