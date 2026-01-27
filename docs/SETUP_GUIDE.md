# Setup Complete! Next Steps

セットアップが完了しました。以下の手順に従って設定を行ってください。

## 1. 設定ファイルの編集

`config.yaml` を編集し、以下の項目を設定してください：

- **repository**: 対象の GitHub リポジトリ名 (例: `owner/repo`)
- **channel**: 通知先の Slack チャンネル (例: `#dev-team`)
- **mapping**: GitHub ユーザー名と Slack Member ID のマッピング

## 2. GitHub Secrets の設定

リポジトリの **Settings** > **Secrets and variables** > **Actions** に以下の Secret を追加してください：

| Secret Name | Value | Description |
|---|---|---|
| `GH_PAT` | GitHub Personal Access Token | プライベートリポジトリの読み取り権限が必要です ([作成方法](https://github.com/settings/tokens)). Scope: `repo` |
| `SLACK_TOKEN` | Slack Bot User OAuth Token | Slack アプリのトークン ([詳細](https://api.slack.com/apps)). Scope: `chat:write` |

## 3. ファイルのコミットとプッシュ

追加・変更されたファイルをコミットし、プッシュしてください。

```bash
git add .
git commit -m "Setup Slack PR Reminder"
git push
```

## 4. 動作確認

GitHub の **Actions** タブから **PR Reminder** ワークフローを手動実行し、動作およびログを確認してください。
