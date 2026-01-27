# Slack PR Reminder

GitHubのOpenなPull Requestに対し、未レビューのAssigneeへSlackでリマインドを送るツールです。

## セットアップ

### 前提条件

- GitHub リポジトリへのアクセス権限
- Slack ワークスペースの管理者権限（Bot作成のため）
- GitHub Actions が有効なリポジトリ

### インストール手順

#### 1. ツールコードの配置

このディレクトリ一式を、対象リポジトリの `tools/pr_reminder/` に配置してください。

```bash
# 例: 対象リポジトリのルートで実行
mkdir -p tools/pr_reminder
# このリポジトリの内容を tools/pr_reminder/ にコピー
```

#### 2. セットアップスクリプトの実行

配置したディレクトリで `setup.sh` を実行します。

```bash
cd tools/pr_reminder
chmod +x setup.sh
./setup.sh
```

スクリプトが以下の処理を自動で行います：
- `config.yaml` をテンプレートからコピー
- `.github/workflows/pr-reminder.yml` を配置
- ワークフロー内のパスを自動調整

#### 3. 設定ファイルの編集

`tools/pr_reminder/config.yaml` を編集し、以下の項目を設定してください：

```yaml
github:
  repository: "owner/repo"  # 対象リポジトリ

slack:
  channel: "#your-channel"  # 通知先チャンネル
  mapping:
    "github_user1": "U12345678"  # GitHubユーザー名: Slack Member ID
```

**Slack Member ID の確認方法:**
1. Slack でユーザーのプロフィールを開く
2. 「その他」→「メンバーIDをコピー」

#### 4. GitHub Secrets の設定

リポジトリの **Settings** > **Secrets and variables** > **Actions** で以下を設定：

| Secret 名 | 説明 | 取得方法 |
|-----------|------|----------|
| `GH_PAT` | GitHub Personal Access Token | [GitHub Settings](https://github.com/settings/tokens) で作成。Scope: `repo` |
| `SLACK_TOKEN` | Slack Bot Token | [Slack API](https://api.slack.com/apps) でアプリ作成。Scope: `chat:write` |

#### 5. コミット & プッシュ

```bash
git add .
git commit -m "Add Slack PR Reminder"
git push
```

#### 6. 動作確認

GitHub の **Actions** タブから **PR Reminder** ワークフローを手動実行し、動作を確認してください。

## 機能

- GitHubリポジトリからOpen状態のPRを取得
- PR作成後1時間以上経過したPRは、毎時未レビュー者にリマインドを送信
- 未レビューのAssigneeに通知（レビュー済みのユーザーはスキップ）
- 営業時間内（平日10:00-19:00 JST）のみ動作
- 日本の祝日・年末年始休暇に対応

## 自動実行スケジュール

デフォルトでは、平日 10:00〜19:00（JST）の毎時0分に自動実行されます。

スケジュールを変更したい場合は、`.github/workflows/pr-reminder.yml` の `cron` 設定を編集してください。

## トラブルシューティング

### リマインダーが送信されない

1. GitHub Actions のログを確認
2. `config.yaml` の設定を確認（リポジトリ名、チャンネル名、マッピング）
3. GitHub Secrets が正しく設定されているか確認
4. 営業時間内に実行されているか確認

### Slack に通知が届かない

1. Slack Bot Token の権限を確認（`chat:write` が必要）
2. Bot がチャンネルに追加されているか確認
3. チャンネル名が正しいか確認（`#` を含む）

## 詳細ドキュメント

- [開発者向けドキュメント](src/README.md) - 内部実装や開発に関する情報
- [セットアップガイド](docs/SETUP_GUIDE.md) - 詳細なセットアップ手順
- [GitHub Token 取得方法](docs/github-token-setup.md)
- [Cron 設定ガイド](docs/cron-setup.md) - ローカル/サーバーでの定期実行

## ライセンス

MIT License
