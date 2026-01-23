# Slack PR Reminder

GitHubのOpenなPull Requestに対し、未レビューのAssigneeへSlackでリマインドを送るツールです。

GitHub Actionsで自動実行されます。

## 機能

- GitHubリポジトリからOpen状態のPRを取得
- PR作成から1時間ごとにリマインダーを送信
- 未レビューのAssigneeのみに通知（レビュー済みのユーザーはスキップ）
- 営業時間内（平日10:00-19:00 JST）のみ動作
- 日本の祝日・年末年始休暇に対応

## クイックスタート

### 1. リポジトリをフォーク/クローン

```bash
git clone https://github.com/daikikim/slack-pr-reminder.git
cd slack-pr-reminder
```

### 2. 設定ファイルを編集

`config.yaml` を編集して、対象リポジトリとSlack設定を入力：

```yaml
github:
  repository: "owner/repo"  # 対象リポジトリ

slack:
  channel: "#pr-reminders"  # 通知先チャンネル
  mapping:                  # GitHubユーザー名 → Slack ID
    "github_user1": "U12345678"
    "github_user2": "U87654321"
```

### 3. GitHub Secrets を設定

リポジトリの **Settings → Secrets and variables → Actions** で以下を設定：

| Secret 名 | 説明 |
|-----------|------|
| `GH_PAT` | GitHub Personal Access Token（[取得方法](docs/github-token-setup.md)） |
| `SLACK_TOKEN` | Slack Bot Token |

### 4. 自動実行開始

設定をプッシュすると、GitHub Actionsが平日10:00〜19:00（JST）に毎時自動実行されます。

**手動実行でテスト:**
1. GitHub リポジトリの **Actions** タブを開く
2. **PR Reminder** ワークフローを選択
3. **Run workflow** をクリック

## 設定

### 環境変数

| 変数名 | 説明 | 必須 |
|--------|------|------|
| `GITHUB_TOKEN` | GitHub Personal Access Token | Yes |
| `SLACK_TOKEN` | Slack Bot Token | Yes |

### GitHub Token

以下の権限が必要です:
- `repo` - プライベートリポジトリの場合
- `public_repo` - パブリックリポジトリのみの場合

詳細な取得手順は [GitHub Token 取得ガイド](docs/github-token-setup.md) を参照してください。

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
    - "2025-01-04"
  new_year_break:           # 年末年始休暇（MM-DD形式）
    start: "12-29"
    end: "01-03"
```

## GitHub Actions

### 自動実行スケジュール

`.github/workflows/pr-reminder.yml` で定義：

- **実行タイミング**: 平日 10:00〜19:00（JST）の毎時0分
- **手動実行**: Actions タブから随時実行可能

### ワークフロー設定

```yaml
name: PR Reminder

on:
  schedule:
    - cron: '0 1-10 * * 1-5'  # UTC時間（JST 10:00-19:00）
  workflow_dispatch:          # 手動実行

jobs:
  remind:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'
      - run: go build -o pr-reminder ./src/cmd/pr-reminder
      - run: ./pr-reminder -config config.yaml
        env:
          GITHUB_TOKEN: ${{ secrets.GH_PAT }}
          SLACK_TOKEN: ${{ secrets.SLACK_TOKEN }}
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

## ローカル実行（オプション）

GitHub Actions を使わずにローカルで実行することも可能です。

### ビルド

```bash
go build -o pr-reminder ./src/cmd/pr-reminder
```

### 実行

```bash
# .env ファイルを作成
cp .env.example .env
vim .env  # トークンを設定

# 実行
./pr-reminder -config config.yaml

# Dry-run（Slack送信なし）
./pr-reminder -dry-run -config config.yaml
```

### コマンドラインオプション

| オプション | デフォルト | 説明 |
|-----------|-----------|------|
| `-config` | `config.yaml` | 設定ファイルのパス |
| `-env` | `.env` | 環境変数ファイルのパス |
| `-dry-run` | `false` | Dry-runモード（Slack送信なし） |

### Cron での定期実行

ローカルマシンやサーバーで定期実行する場合は [Cron 設定ガイド](docs/cron-setup.md) を参照してください。

## ディレクトリ構成

```
slack-pr-reminder/
├── .github/workflows/           # GitHub Actions
│   └── pr-reminder.yml
├── src/
│   ├── cmd/pr-reminder/         # エントリーポイント
│   └── internal/
│       ├── model/               # エンティティ・インターフェース
│       ├── view/                # Slackメッセージ生成
│       ├── controller/          # ビジネスロジック
│       └── infrastructure/      # 外部サービス連携
├── docs/                        # ドキュメント
├── config.yaml                  # 設定ファイル
└── .env.example                 # 環境変数サンプル
```

## 開発

### テスト実行

```bash
go test ./... -v
```

### ビルド

```bash
go build -o pr-reminder ./src/cmd/pr-reminder
```

## 依存ライブラリ

- [google/go-github](https://github.com/google/go-github) - GitHub API クライアント
- [slack-go/slack](https://github.com/slack-go/slack) - Slack API クライアント
- [holiday-jp/holiday_jp-go](https://github.com/holiday-jp/holiday_jp-go) - 日本の祝日判定
- [joho/godotenv](https://github.com/joho/godotenv) - .env ファイル読み込み
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3) - YAML パーサー

## ライセンス

MIT License
