#!/bin/bash

# setup.sh
# Slack PR Reminder セットアップスクリプト
# このスクリプトは、ツールコード一式が配置されたディレクトリ（例: tools/pr_reminder）で実行してください。

set -e

# 現在のディレクトリを取得
CURRENT_DIR=$(pwd)
REPO_ROOT=$(git rev-parse --show-toplevel)
WORKFLOW_DIR="$REPO_ROOT/.github/workflows"

echo "=== Slack PR Reminder セットアップ ==="
echo "現在のディレクトリ: $CURRENT_DIR"
echo "リポジトリルート:   $REPO_ROOT"
echo ""

# 1. config.yaml のセットアップ
if [ -f "config.yaml" ]; then
    echo "[スキップ] config.yaml は既に存在します"
else
    echo "[コピー] テンプレートから config.yaml を作成しています..."
    cp templates/config.yaml config.yaml
    echo "完了"
fi

# 2. GitHub Actions ワークフローのセットアップ
mkdir -p "$WORKFLOW_DIR"

WORKFLOW_DEST="$WORKFLOW_DIR/pr-reminder.yml"
if [ -f "$WORKFLOW_DEST" ]; then
    echo "[スキップ] ワークフローファイルは既に存在します: $WORKFLOW_DEST"
else
    echo "[コピー] ワークフローファイルを作成しています: $WORKFLOW_DEST..."
    cp templates/pr-reminder.yml "$WORKFLOW_DEST"
    echo "完了"
fi

# 3. ワークフロー内のパス調整
# このスクリプトが実行されているディレクトリ（リポジトリルートからの相対パス）を取得
RELATIVE_PATH=${CURRENT_DIR#$REPO_ROOT/}

if [ "$RELATIVE_PATH" == "$CURRENT_DIR" ]; then
    # ルートディレクトリで実行されている場合
    RELATIVE_PATH="."
fi

echo "[更新] ワークフローファイルの working-directory を更新しています: $RELATIVE_PATH"

# OSによってsedの挙動が異なるため分岐
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s|working-directory: .*|working-directory: $RELATIVE_PATH|g" "$WORKFLOW_DEST"
else
    # Linux
    sed -i "s|working-directory: .*|working-directory: $RELATIVE_PATH|g" "$WORKFLOW_DEST"
fi
echo "完了"

# 4. セットアップガイドの表示
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SETUP_GUIDE="$SCRIPT_DIR/docs/SETUP_GUIDE.md"

if [ -f "$SETUP_GUIDE" ]; then
    echo ""
    echo "=== セットアップが正常に完了しました ==="
    echo ""
    cat "$SETUP_GUIDE"
else
    echo "[警告] SETUP_GUIDE.md が見つかりません: $SETUP_GUIDE"
    echo "次のステップについてはドキュメントを参照してください"
fi
