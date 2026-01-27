# Cron 設定ガイド

このドキュメントでは、Cron を使用して Slack PR Reminder を定期実行する方法を説明します。

## 実行環境の選択

Slack PR Reminder を定期実行する方法は、実行環境によって異なります：

### 1. ローカルマシン（このドキュメントの主な内容）

**メリット:**
- 設定が簡単
- 追加のインフラ不要
- 開発・テストに適している

**デメリット:**
- マシンが常時起動している必要がある
- マシンがスリープすると実行されない
- 個人のマシンに依存する

**推奨用途:** 個人利用、開発・テスト環境

### 2. サーバー（Linux サーバー、VPS、EC2 など）

**メリット:**
- 24時間稼働が可能
- 安定した実行環境
- 複数のサービスを統合管理可能

**デメリット:**
- サーバーの維持・管理が必要
- コストがかかる場合がある

**推奨用途:** 本番環境、チーム利用

### 3. クラウドサービス（GitHub Actions、AWS Lambda など）

**メリット:**
- サーバー管理が不要
- スケーラブル
- 多くの場合、無料枠がある

**デメリット:**
- 設定がやや複雑
- サービス固有の制限がある

**推奨用途:** 本番環境、CI/CD パイプラインとの統合

> **注意:** このドキュメントでは主にローカルマシンとサーバーでの Cron 設定を説明しています。クラウドサービスでの実行方法については、各サービスのドキュメントを参照してください。

## Cron とは

Cron は Unix/Linux 系 OS に標準で搭載されているタスクスケジューラです。指定した時間に自動でコマンドを実行できます。

**主な用途:**
- 定期的なバックアップ
- ログのローテーション
- 定期的なデータ取得・通知（← 今回の用途）

## Cron の基本

### Crontab（Cron Table）

Cron の設定は「crontab」というファイルで管理します。

```bash
# 現在の設定を表示
crontab -l

# 設定を編集
crontab -e

# 設定を削除（注意！）
crontab -r
```

### Cron 式の書き方

Cron 式は5つのフィールドで構成されます：

```
┌───────────── 分 (0-59)
│ ┌───────────── 時 (0-23)
│ │ ┌───────────── 日 (1-31)
│ │ │ ┌───────────── 月 (1-12)
│ │ │ │ ┌───────────── 曜日 (0-7, 0と7は日曜日)
│ │ │ │ │
│ │ │ │ │
* * * * * コマンド
```

### 特殊文字

| 文字 | 意味 | 例 |
|------|------|-----|
| `*` | すべての値 | `* * * * *` = 毎分 |
| `,` | 複数の値 | `0,30 * * * *` = 毎時0分と30分 |
| `-` | 範囲 | `0 9-17 * * *` = 9時〜17時の毎時0分 |
| `/` | 間隔 | `*/15 * * * *` = 15分ごと |

### よく使う Cron 式の例

| Cron 式 | 説明 |
|---------|------|
| `0 * * * *` | 毎時0分 |
| `*/30 * * * *` | 30分ごと |
| `0 9 * * *` | 毎日 9:00 |
| `0 9 * * 1-5` | 平日 9:00 |
| `0 10-18 * * 1-5` | 平日 10:00〜18:00 の毎時0分 |
| `0 9 1 * *` | 毎月1日 9:00 |
| `0 0 * * 0` | 毎週日曜 0:00 |

## PR Reminder の Cron 設定

### 推奨設定

PR Reminder は営業時間内（10:00-19:00）のみ動作するため、その時間帯で実行すれば十分です。

```bash
# 平日 10:00〜19:00 の毎時0分に実行
0 10-19 * * 1-5 /path/to/pr-reminder -config /path/to/config.yaml
```

### 設定手順

#### 1. バイナリのパスを確認

```bash
# バイナリの絶対パスを取得
cd /path/to/slack-pr-reminder
pwd
# 例: /Users/username/slack-pr-reminder
```

#### 2. Crontab を編集

```bash
crontab -e
```

#### 3. 設定を追加

```cron
# Slack PR Reminder - 平日 10:00〜19:00 に毎時実行
0 10-19 * * 1-5 cd /Users/username/slack-pr-reminder && ./pr-reminder >> /tmp/pr-reminder.log 2>&1
```

**ポイント:**
- `cd /path/to/project &&` で作業ディレクトリを移動（.env と config.yaml を読み込むため）
- `>> /tmp/pr-reminder.log 2>&1` でログをファイルに保存
- `2>&1` で標準エラー出力も標準出力にリダイレクト

#### 4. 設定を確認

```bash
crontab -l
```

### 環境変数の注意点

Cron は通常のシェルとは異なる環境で実行されます。環境変数が読み込まれない場合があります。

**解決方法 1: .env ファイルを使用（推奨）**

PR Reminder は `.env` ファイルをサポートしているため、トークンを `.env` に記載すれば問題ありません。

**解決方法 2: Crontab で環境変数を設定**

```cron
GITHUB_TOKEN=ghp_xxxxxxxxxxxx
SLACK_TOKEN=xoxb-xxxxxxxxxxxx

0 10-19 * * 1-5 cd /path/to/project && ./pr-reminder
```

**解決方法 3: スクリプトでラップ**

```bash
#!/bin/bash
# run-pr-reminder.sh

export GITHUB_TOKEN="ghp_xxxxxxxxxxxx"
export SLACK_TOKEN="xoxb-xxxxxxxxxxxx"

cd /path/to/slack-pr-reminder
./pr-reminder -config config.yaml
```

```cron
0 10-19 * * 1-5 /path/to/run-pr-reminder.sh >> /tmp/pr-reminder.log 2>&1
```

## サーバー環境での設定

### Linux サーバー（VPS、EC2 など）

サーバー環境では、通常の Cron 設定と同じ手順で設定できます。

#### 1. バイナリをサーバーに配置

```bash
# サーバーに SSH 接続
ssh user@your-server.com

# プロジェクトディレクトリを作成
mkdir -p /opt/slack-pr-reminder
cd /opt/slack-pr-reminder

# バイナリと設定ファイルを配置
# (git clone または scp で転送)
```

#### 2. 環境変数と設定ファイルを配置

```bash
# .env ファイルを作成
vim .env

# config.yaml を配置
vim config.yaml
```

#### 3. Crontab を設定

```bash
# root ユーザーまたは専用ユーザーで設定
sudo crontab -e

# または、専用ユーザーを作成
sudo useradd -m -s /bin/bash pr-reminder
sudo su - pr-reminder
crontab -e
```

```cron
# Slack PR Reminder - 平日 10:00〜19:00 に毎時実行
0 10-19 * * 1-5 cd /opt/slack-pr-reminder && ./pr-reminder -config config.yaml >> /var/log/pr-reminder.log 2>&1
```

#### 4. ログローテーションの設定（推奨）

```bash
# /etc/logrotate.d/pr-reminder を作成
sudo vim /etc/logrotate.d/pr-reminder
```

```
/var/log/pr-reminder.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
}
```

### クラウドサービスでの実行

#### GitHub Actions（推奨）

GitHub Actions を使用すると、サーバー管理不要で定期実行できます。

##### セットアップ手順

**1. GitHub Secrets を設定**

リポジトリの Settings → Secrets and variables → Actions で以下を設定：

| Secret 名 | 値 | 説明 |
|-----------|-----|------|
| `GH_PAT` | Personal Access Token | PR 情報取得用（[取得方法](github-token-setup.md)） |
| `SLACK_TOKEN` | Slack Bot Token | Slack 通知用 |

> **注意:** デフォルトの `GITHUB_TOKEN` は権限が限定されているため、別途 Personal Access Token (PAT) を `GH_PAT` として設定する必要があります。

**2. ワークフローファイル**

`.github/workflows/pr-reminder.yml` が既に配置されています：

```yaml
name: PR Reminder

on:
  schedule:
    # 平日 10:00〜19:00 の毎時0分（JST）
    # JST 10:00 = UTC 01:00, JST 19:00 = UTC 10:00
    - cron: '0 1-10 * * 1-5'
  workflow_dispatch: # 手動実行も可能

jobs:
  remind:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'

      - name: Build
        run: go build -o pr-reminder ./src/cmd/pr-reminder

      - name: Run PR Reminder
        env:
          GITHUB_TOKEN: ${{ secrets.GH_PAT }}
          SLACK_TOKEN: ${{ secrets.SLACK_TOKEN }}
        run: ./pr-reminder -config config.yaml
```

**3. 手動実行でテスト**

1. GitHub リポジトリの Actions タブを開く
2. 「PR Reminder」ワークフローを選択
3. 「Run workflow」ボタンをクリック

**メリット:**
- サーバー管理不要
- GitHub リポジトリと統合
- 無料枠あり（月2,000分）
- 手動実行でテスト可能

#### AWS Lambda + EventBridge

AWS Lambda を使用する場合は、Go の Lambda ランタイムを使用します。

## macOS での設定

### launchd（macOS 推奨）

macOS では Cron の代わりに launchd を使用することも推奨されています。

#### 1. plist ファイルを作成

```bash
vim ~/Library/LaunchAgents/com.user.pr-reminder.plist
```

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.user.pr-reminder</string>

    <key>ProgramArguments</key>
    <array>
        <string>/Users/username/slack-pr-reminder/pr-reminder</string>
        <string>-config</string>
        <string>/Users/username/slack-pr-reminder/config.yaml</string>
        <string>-env</string>
        <string>/Users/username/slack-pr-reminder/.env</string>
    </array>

    <key>WorkingDirectory</key>
    <string>/Users/username/slack-pr-reminder</string>

    <key>StartCalendarInterval</key>
    <array>
        <!-- 平日 10:00〜19:00 の毎時0分 -->
        <dict>
            <key>Weekday</key><integer>1</integer>
            <key>Hour</key><integer>10</integer>
            <key>Minute</key><integer>0</integer>
        </dict>
        <dict>
            <key>Weekday</key><integer>1</integer>
            <key>Hour</key><integer>11</integer>
            <key>Minute</key><integer>0</integer>
        </dict>
        <!-- ... 他の時間も同様に追加 ... -->
    </array>

    <key>StandardOutPath</key>
    <string>/tmp/pr-reminder.log</string>

    <key>StandardErrorPath</key>
    <string>/tmp/pr-reminder.error.log</string>
</dict>
</plist>
```

#### 2. 読み込み・有効化

```bash
# 読み込み
launchctl load ~/Library/LaunchAgents/com.user.pr-reminder.plist

# 無効化
launchctl unload ~/Library/LaunchAgents/com.user.pr-reminder.plist

# 状態確認
launchctl list | grep pr-reminder
```

### macOS で Cron を使う場合

macOS でも Cron は使用可能ですが、「フルディスクアクセス」の許可が必要な場合があります。

```
システム設定 → プライバシーとセキュリティ → フルディスクアクセス → cron を追加
```

## トラブルシューティング

### Cron が実行されない

1. **Cron サービスの確認**
   ```bash
   # Linux
   sudo systemctl status cron

   # macOS
   sudo launchctl list | grep cron
   ```

2. **パスの確認**
   - 絶対パスを使用しているか確認
   - バイナリに実行権限があるか確認: `chmod +x pr-reminder`

3. **ログの確認**
   ```bash
   # システムログ
   grep CRON /var/log/syslog  # Linux
   log show --predicate 'process == "cron"' --last 1h  # macOS

   # アプリケーションログ
   tail -f /tmp/pr-reminder.log
   ```

### 環境変数が読み込まれない

- `.env` ファイルのパスが正しいか確認
- 作業ディレクトリを `cd` で移動しているか確認

### タイムゾーンの問題

Cron はシステムのタイムゾーンで動作します。

```bash
# タイムゾーン確認
date +%Z

# タイムゾーン設定（Linux）
sudo timedatectl set-timezone Asia/Tokyo
```

## 動作確認

### 手動テスト

```bash
# Dry-run で動作確認
cd /path/to/slack-pr-reminder
./pr-reminder -dry-run
```

### Cron の即時テスト

一時的に毎分実行する設定を追加して確認：

```cron
# テスト用（確認後に削除）
* * * * * cd /path/to/project && ./pr-reminder -dry-run >> /tmp/pr-reminder-test.log 2>&1
```

```bash
# ログを監視
tail -f /tmp/pr-reminder-test.log
```

## 参考リンク

- [Crontab Guru](https://crontab.guru/) - Cron 式のビジュアルエディタ
- [crontab.5 man page](https://man7.org/linux/man-pages/man5/crontab.5.html)
- [Apple Developer - launchd](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/ScheduledJobs.html)
