# TablePlus で minikube PostgreSQL に接続するガイド

## 概要

このドキュメントでは、GUI データベースクライアントの TablePlus を使用して、minikube 内で動作している PostgreSQL データベースに接続する手順を説明します。

## 前提条件

- TablePlus がインストールされていること
- minikube が起動していること
- kubectl が設定されていること
- PostgreSQL Pod が稼働していること

## 接続情報

| 項目 | 値 |
|------|-----|
| ホスト | localhost または 127.0.0.1 |
| ポート | 5432 |
| ユーザー名 | postgres |
| パスワード | postgres123 |
| データベース | postgresdb |

## 接続手順

### 1. ポートフォワードの設定

まず、kubectl を使用して PostgreSQL Pod のポートをローカルに転送します：

```bash
kubectl port-forward postgres-0 5432:5432
```

このコマンドは接続を維持するために実行し続ける必要があります。以下のような出力が表示されます：

```
Forwarding from 127.0.0.1:5432 -> 5432
Forwarding from [::1]:5432 -> 5432
```

### 2. TablePlus での接続設定

1. TablePlus を起動します
2. 「Create a new connection」をクリック
3. 「PostgreSQL」を選択
4. 以下の情報を入力します：

![TablePlus 接続設定画面](connection-settings.png)

```
Name: minikube-postgres（任意の名前）
Host: localhost
Port: 5432
User: postgres
Password: postgres123
Database: postgresdb
```

5. 「Test」ボタンをクリックして接続を確認
6. 「Connect」をクリックして接続

### 3. 接続確認

接続が成功すると、データベースの内容が表示されます。

## トラブルシューティング

### 接続できない場合

1. **ポートフォワードが動作しているか確認**
   ```bash
   # 別のターミナルで実行
   lsof -i :5432
   ```

2. **Pod が稼働しているか確認**
   ```bash
   kubectl get pod postgres-0
   ```

3. **既存の PostgreSQL プロセスがポートを使用していないか確認**
   ```bash
   # macOS の場合
   brew services list | grep postgresql

   # 停止する場合
   brew services stop postgresql
   ```

### よくあるエラー

#### "connection refused" エラー
- ポートフォワードが実行されていない可能性があります
- ターミナルでポートフォワードコマンドを再実行してください

#### "password authentication failed" エラー
- パスワードが正しいか確認してください
- Secret の値を再確認：
  ```bash
  kubectl get secret postgres-secret -o jsonpath='{.data.POSTGRES_PASSWORD}' | base64 -d
  ```

## セキュリティに関する注意事項

- この接続方法は開発環境向けです
- 本番環境では以下を検討してください：
  - SSH トンネル
  - VPN 接続
  - 専用のデータベースプロキシ
  - より強固な認証方法

## 便利な使い方

### SQL クエリの実行

TablePlus では以下の機能が使用できます：

1. **クエリエディタ**: Cmd+T (Mac) / Ctrl+T (Windows) で新しいクエリタブを開く
2. **テーブル構造の確認**: テーブルを右クリックして「Structure」を選択
3. **データのエクスポート**: テーブルを右クリックして「Export」を選択
4. **インポート**: File メニューから「Import」を選択

### よく使うクエリ

```sql
-- データベースサイズの確認
SELECT pg_database_size(current_database()) / 1024 / 1024 as size_mb;

-- テーブル一覧
SELECT tablename FROM pg_tables WHERE schemaname = 'public';

-- アクティブな接続数
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';
```

## まとめ

kubectl port-forward を使用することで、簡単に minikube 内の PostgreSQL データベースに TablePlus から接続できます。この方法は開発やデバッグに便利ですが、常にセキュリティを考慮して使用してください。