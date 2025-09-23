FROM golang:1.21-alpine AS builder

WORKDIR /app

# 依存関係のインストール
RUN apk add --no-cache git

# go.mod と go.sum をコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードをコピー
COPY main.go .

# ビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slack-metrics-api .

# 実行用の軽量イメージ
FROM alpine:3.18

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# ビルドしたバイナリをコピー
COPY --from=builder /app/slack-metrics-api .

# 実行権限を付与
RUN chmod +x ./slack-metrics-api

# ポート公開
EXPOSE 8080

CMD ["./slack-metrics-api"]
