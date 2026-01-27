# Slack PR Reminder - 開発者向けドキュメント

このドキュメントは、Slack PR Reminder の内部実装や開発に関する情報を記載しています。

## アーキテクチャ

### ディレクトリ構成

```
src/
├── cmd/pr-reminder/         # エントリーポイント
│   └── main.go
└── internal/
    ├── model/               # エンティティ・インターフェース
    │   ├── entity.go
    │   └── interface.go
    ├── view/                # Slackメッセージ生成
    │   └── slack_view.go
    ├── controller/          # ビジネスロジック
    │   ├── reminder_controller.go
    │   └── reminder_controller_test.go
    └── infrastructure/      # 外部サービス連携
        ├── config/          # 設定ファイル読み込み
        ├── github/          # GitHub API クライアント
        ├── slack/           # Slack API クライアント
        └── time/            # 営業時間判定
```

### 主要コンポーネント

#### ReminderController
PR リマインダーのメインロジックを制御します。

- 営業時間のチェック
- Open PR の取得
- レビュー状況の確認
- バッチリマインダーの送信

#### PRRepository (GitHub Client)
GitHub API を使用して PR 情報を取得します。

- `FetchOpenPRs`: Open 状態の PR 一覧を取得
- `GetReviewStatuses`: PR のレビュー状況を取得
- `IsMerged`: PR がマージ済みかチェック

#### Notifier (Slack Client)
Slack API を使用してメッセージを送信します。

- `SendReminder`: 指定チャンネルにメッセージを送信

#### TimeChecker (BusinessTimer)
営業時間や祝日を判定します。

- `IsBusinessTime`: 現在時刻が営業時間内かチェック
- `ShouldRemind`: リマインダーを送信すべきタイミングかチェック

## リマインダーの動作

### 送信タイミング

- PR作成後1時間以上経過したPRは、毎時（GitHub Actionsの実行タイミングで）未レビュー者にリマインドを送信
- 実行頻度はCronの設定に依存します（通常は1時間に1回）

### 送信条件

以下の条件をすべて満たす場合にリマインダーを送信:

1. 現在時刻が営業時間内（平日10:00-19:00）
2. 祝日・年末年始でない
3. PR作成から1時間以上経過している
4. 未レビューのAssigneeが存在する

### メッセージ形式

複数のPRをまとめて1つのメッセージで通知:

```
PRのレビューをお願いします

https://github.com/owner/repo/pull/123
・対象者：<@user1>
・経過時間：2時間

https://github.com/owner/repo/pull/124
・対象者：<@user2>, <@user3>
・経過時間：5時間
```

## 開発

### 必要な環境

- Go 1.21 以上
- GitHub Personal Access Token (scope: `repo`)
- Slack Bot Token (scope: `chat:write`)

### ローカル実行

#### ビルド

```bash
go build -o pr-reminder ./src/cmd/pr-reminder
```

#### 実行

```bash
# .env ファイルを作成
cp .env.example .env
vim .env  # トークンを設定

# 実行
./pr-reminder -config config.yaml

# Dry-run（Slack送信なし）
./pr-reminder -dry-run -config config.yaml
```

#### コマンドラインオプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `-config` | `config.yaml` | 設定ファイルのパス |
| `-env` | `.env` | 環境変数ファイルのパス |
| `-dry-run` | `false` | Dry-runモード（Slack送信なし） |

### テスト実行

```bash
go test ./... -v
```

### 設定ファイル (config.yaml)

```yaml
github:
  repository: "owner/repo"  # 対象リポジトリ（owner/repo形式）

slack:
  channel: "#pr-reminders"  # 通知先チャンネル
  mapping:                  # GitHubユーザー名 → Slack ID のマッピング
    "github_user1": "U12345678"
    "github_user2": "U87654321"

schedule:
  business_hours:
    start: "10:00"          # 営業開始時刻
    end: "19:00"            # 営業終了時刻
  timezone: "Asia/Tokyo"    # タイムゾーン
  holidays:                 # 追加の休日（YYYY-MM-DD形式）
    - "2025-01-04"
  new_year_break:           # 年末年始休暇（MM-DD形式）
    start: "12-29"
    end: "01-03"
```

## 依存ライブラリ

- [google/go-github](https://github.com/google/go-github) - GitHub API クライアント
- [slack-go/slack](https://github.com/slack-go/slack) - Slack API クライアント
- [holiday-jp/holiday_jp-go](https://github.com/holiday-jp/holiday_jp-go) - 日本の祝日判定
- [joho/godotenv](https://github.com/joho/godotenv) - .env ファイル読み込み
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML パーサー

## ライセンス

MIT License
