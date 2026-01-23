# GitHub Token 取得ガイド

このドキュメントでは、Slack PR Reminder で使用する GitHub Personal Access Token (PAT) の取得方法を説明します。

## Personal Access Token の種類

GitHub には2種類の Personal Access Token があります:

| 種類 | 説明 | 推奨 |
|------|------|------|
| **Fine-grained PAT** | リポジトリ単位で細かく権限設定可能 | ✅ 推奨 |
| **Classic PAT** | 従来の方式、広範囲な権限 | セキュリティ上非推奨 |

## Fine-grained Personal Access Token の作成（推奨）

### 手順

1. **GitHub にログイン**
   - https://github.com にアクセスしてログイン

2. **Settings を開く**
   - 右上のプロフィールアイコンをクリック
   - 「Settings」を選択

3. **Developer settings に移動**
   - 左サイドバーの一番下にある「Developer settings」をクリック

4. **Personal access tokens を選択**
   - 「Personal access tokens」を展開
   - 「Fine-grained tokens」をクリック

5. **新しいトークンを生成**
   - 「Generate new token」ボタンをクリック

6. **トークン設定を入力**

   | 項目 | 設定値 |
   |------|--------|
   | Token name | `slack-pr-reminder` (任意の名前) |
   | Expiration | 90 days〜1 year (推奨: 90 days) |
   | Description | PR reminder tool for Slack notifications |
   | Resource owner | 対象リポジトリのオーナー（個人 or Organization） |
   | Repository access | 「Only select repositories」を選択し、対象リポジトリを選択 |

7. **権限を設定**

   「Repository permissions」セクションで以下を設定:

   | 権限 | レベル | 用途 |
   |------|--------|------|
   | **Pull requests** | Read-only | PR一覧の取得 |
   | **Metadata** | Read-only | リポジトリ情報の取得（自動で付与） |

8. **トークンを生成**
   - 「Generate token」ボタンをクリック
   - 表示されたトークンをコピー

   ⚠️ **重要**: トークンは一度しか表示されません。必ずコピーして安全な場所に保存してください。

### トークンの保存

```bash
# .env ファイルに保存
echo "GITHUB_TOKEN=github_pat_xxxxxxxxxxxxxx" >> .env
```

## Classic Personal Access Token の作成（非推奨）

Fine-grained PAT が使用できない場合のみ、こちらの方法を使用してください。

### 手順

1. **GitHub にログイン**

2. **Settings → Developer settings → Personal access tokens → Tokens (classic)**

3. **「Generate new token (classic)」をクリック**

4. **設定を入力**

   | 項目 | 設定値 |
   |------|--------|
   | Note | `slack-pr-reminder` |
   | Expiration | 90 days (推奨) |

5. **スコープを選択**

   | スコープ | 説明 |
   |----------|------|
   | `repo` | プライベートリポジトリにアクセスする場合 |
   | `public_repo` | パブリックリポジトリのみの場合 |

6. **「Generate token」をクリック**

## トークンの確認

トークンが正しく動作するか確認:

```bash
# トークンをエクスポート
export GITHUB_TOKEN=your_token_here

# API をテスト
curl -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://api.github.com/repos/OWNER/REPO/pulls
```

成功すると、PR一覧の JSON が返されます。

## トラブルシューティング

### 401 Unauthorized エラー

- トークンが正しくコピーされているか確認
- トークンの有効期限が切れていないか確認
- 環境変数が正しく設定されているか確認:
  ```bash
  echo $GITHUB_TOKEN
  ```

### 403 Forbidden エラー

- トークンに必要な権限があるか確認
- Fine-grained PAT の場合、対象リポジトリが選択されているか確認
- Organization リポジトリの場合、Organization が Fine-grained PAT を許可しているか確認

### 404 Not Found エラー

- リポジトリ名が正しいか確認（`owner/repo` 形式）
- プライベートリポジトリの場合、`repo` スコープがあるか確認

## セキュリティのベストプラクティス

1. **最小権限の原則**
   - 必要最小限の権限のみ付与
   - Fine-grained PAT を使用して対象リポジトリを限定

2. **有効期限の設定**
   - 90日など短めの有効期限を設定
   - 期限が近づいたら新しいトークンを発行

3. **トークンの保管**
   - `.env` ファイルは `.gitignore` に追加済み
   - 本番環境ではシークレット管理サービスを使用

4. **定期的なローテーション**
   - 定期的にトークンを再発行
   - 不要になったトークンは削除

## 参考リンク

- [GitHub Docs: Managing your personal access tokens](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)
- [GitHub Docs: Creating a fine-grained personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens#creating-a-fine-grained-personal-access-token)
