# Slack PR Reminder

GitHubのOpenなPull Requestに対し、未レビューのAssigneeへSlackでリマインドを送るツールです。

GitHub Actionsで自動実行されます。

## 機能

- GitHubリポジトリからOpen状態のPRを取得
- PR作成から1時間ごとにリマインダーを送信
- 未レビューのAssigneeに通知（レビュー済みのユーザーはスキップ）
- 全てのAssigneeが承認済み(Approved)の場合、Authorにマージリマインドを送信
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

### セットアップ手順

GitHub Actionsで自動実行するには、以下の手順を完了してください：

#### 1. GitHub Secrets の設定

リポジトリの **Settings** → **Secrets and variables** → **Actions** で以下を設定：

| Secret 名 | 値 | 説明 |
|-----------|-----|------|
| `GH_PAT` | Personal Access Token | PR 情報取得用（[取得方法](docs/github-token-setup.md)） |
| `SLACK_TOKEN` | Slack Bot Token | Slack 通知用 |

> **重要:** デフォルトの `GITHUB_TOKEN` は権限が限定されているため、別途 Personal Access Token (PAT) を `GH_PAT` として設定する必要があります。

#### 2. ワークフローファイルの確認

`.github/workflows/pr-reminder.yml` がデフォルトブランチ（通常は `main` または `master`）に存在することを確認してください。

#### 3. Actions の有効化確認

リポジトリの **Settings** → **Actions** → **General** で以下を確認：
- ✅ Actions permissions が有効になっている
- ✅ Workflow permissions が適切に設定されている

#### 4. 手動実行でテスト

1. GitHub リポジトリの **Actions** タブを開く
2. 左サイドバーで「PR Reminder」ワークフローを選択
3. 右上の「Run workflow」ボタンをクリック
4. 実行結果を確認

### 自動実行スケジュール

`.github/workflows/pr-reminder.yml` で定義：

- **実行タイミング**: 平日 10:00〜19:00（JST）の毎時0分
- **手動実行**: Actions タブから随時実行可能

> **注意:** スケジュール実行は最大で5分程度の遅延が発生する場合があります。また、リポジトリが60日以上非アクティブな場合、スケジュール実行は一時停止されます。

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

- PR作成から1時間以上経過している場合、実行タイミングでリマインダーを送信
- 実行頻度はCronの設定（GitHub Actionsなど）に依存します（通常は1時間に1回）

### 送信条件

以下の条件をすべて満たす場合にリマインダーを送信:

1. 現在時刻が営業時間内（平日10:00-19:00）
2. 祝日・年末年始でない
3. PR作成から1時間以上経過している

**通知の宛先ルール:**

- **レビューリマインド**: 未レビューのAssigneeがいる場合、そのAssignee宛に送信
- **マージリマインド**: Assignee全員が `APPROVED` の場合、Author宛に送信

### メッセージ形式

### レビューリマインド（Reviewer宛）

```
@user PRのレビューをお願いします
**PR Title Here**
https://github.com/owner/repo/pull/123
経過時間: 2時間
```

### マージリマインド（Author宛）

```
@author レビューが全て承認されました！
マージをお願いします
**PR Title Here**
https://github.com/owner/repo/pull/123
経過時間: 5時間
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
