# Slack PR Reminder

GitHubのOpenなPull Requestに対し、未レビューのAssigneeへSlackでリマインドを送るCLIツールです。

## 機能

- GitHubリポジトリからOpen状態のPRを取得
- PR作成から1時間ごとにリマインダーを送信
- 未レビューのAssigneeのみに通知（レビュー済みのユーザーはスキップ）
- 営業時間内（平日10:00-19:00 JST）のみ動作
- 日本の祝日・年末年始休暇に対応
- Dry-runモードでテスト実行可能

## インストール

```bash
go install github.com/dkim/slack-pr-reminder/cmd/pr-reminder@latest
```

または、ソースからビルド:

```bash
git clone https://github.com/dkim/slack-pr-reminder.git
cd slack-pr-reminder
go build -o pr-reminder ./cmd/pr-reminder
```

## 設定

### 環境変数

| 変数名 | 説明 | 必須 |
|--------|------|------|
| `GITHUB_TOKEN` | GitHub Personal Access Token | Yes |
| `SLACK_TOKEN` | Slack Bot Token | Yes (dry-run時は不要) |

環境変数は `.env` ファイルから読み込むことができます:

```bash
# .env ファイルを作成
cp .env.example .env

# .env ファイルを編集
vim .env
```

`.env` ファイルの例:

```env
GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
SLACK_TOKEN=xoxb-xxxxxxxxxxxx-xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxx
```

### GitHub Token

以下の権限が必要です:
- `repo` - プライベートリポジトリの場合
- `public_repo` - パブリックリポジトリのみの場合

### Slack Bot Token

Slack Appを作成し、以下の権限を付与してください:
- `chat:write` - メッセージ送信用

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
    - "2024-01-04"
  new_year_break:           # 年末年始休暇（MM-DD形式）
    start: "12-29"
    end: "01-03"
```

## 使用方法

### 基本的な実行

```bash
# .env ファイルを使用する場合
./pr-reminder -config config.yaml

# または環境変数を直接設定する場合
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
export SLACK_TOKEN=xoxb-xxxxxxxxxxxx
./pr-reminder -config config.yaml
```

### Dry-runモード

Slack APIを呼び出さずに動作確認できます:

```bash
./pr-reminder -dry-run -config config.yaml
```

### コマンドラインオプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `-config` | `config.yaml` | 設定ファイルのパス |
| `-env` | `.env` | 環境変数ファイルのパス |
| `-dry-run` | `false` | Dry-runモード（Slack送信なし） |

## Cron設定例

毎時0分に実行する場合:

```cron
0 * * * * /path/to/pr-reminder -config /path/to/config.yaml >> /var/log/pr-reminder.log 2>&1
```

## リマインダーの動作

### 送信タイミング

- PR作成から1時間経過後、毎時リマインダーを送信
- 例: 10:30に作成されたPR → 11:30, 12:30, 13:30... にリマインド

### 送信条件

以下の条件をすべて満たす場合にリマインダーを送信:

1. 現在時刻が営業時間内（平日10:00-19:00）
2. 祝日・年末年始でない
3. PR作成から1時間以上経過
4. 現在時刻がPR作成時刻の「毎時」のタイミング（±5分の許容範囲）
5. Assigneeがまだレビューしていない
6. AssigneeがPRの作成者ではない

### メッセージ形式

```
@user PRのレビューをお願いします
**PR Title Here**
https://github.com/owner/repo/pull/123
経過時間: 2時間
```

## アーキテクチャ

```
┌─────────────────┐
│   main.go       │  CLI エントリーポイント
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Controller    │  ビジネスロジック
└────────┬────────┘
         │
    ┌────┴────┬──────────┬──────────┐
    ▼         ▼          ▼          ▼
┌───────┐ ┌───────┐ ┌────────┐ ┌───────┐
│GitHub │ │ Slack │ │  Time  │ │ View  │
│Client │ │Client │ │Checker │ │       │
└───────┘ └───────┘ └────────┘ └───────┘
```

## 開発

### テスト実行

```bash
go test ./... -v
```

### ビルド

```bash
go build -o pr-reminder ./cmd/pr-reminder
```

## 依存ライブラリ

- [google/go-github](https://github.com/google/go-github) - GitHub API クライアント
- [slack-go/slack](https://github.com/slack-go/slack) - Slack API クライアント
- [holiday-jp/holiday_jp-go](https://github.com/holiday-jp/holiday_jp-go) - 日本の祝日判定
- [joho/godotenv](https://github.com/joho/godotenv) - .env ファイル読み込み
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML パーサー

## ライセンス

MIT License
