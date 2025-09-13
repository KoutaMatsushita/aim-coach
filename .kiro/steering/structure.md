# File Organization and Code Patterns

## プロジェクト構造

### ディレクトリ構成
```
aim-coach/
├── .kiro/                    # Spec-driven development
│   ├── specs/               # 機能別スペック
│   └── steering/            # プロジェクト全体ガイド
├── .claude/                 # Claude Code設定
├── packages/                # 各スペック実装
│   ├── discord-bot-interface/
│   ├── analytics-engine/
│   ├── data-collector/
│   └── conversation-engine/
├── shared/                  # 共通ユーティリティ
│   ├── types/              # TypeScript型定義
│   ├── utils/              # 共通関数
│   └── constants/          # 定数定義
└── infrastructure/          # デプロイ・設定
    ├── cloudflare/         # Wrangler設定
    ├── database/           # D1スキーマ
    └── deploy/             # デプロイスクリプト
```

## コーディング規約

### TypeScript Standards
- **厳格型チェック**: `strict: true`設定必須
- **命名規約**: camelCase（変数・関数）、PascalCase（クラス・型）
- **インターフェース**: 外部API・内部モジュール間の明確な型定義

### ファイル命名パターン
- **コンポーネント**: `kebab-case.ts`
- **型定義**: `*.types.ts`
- **設定ファイル**: `*.config.ts`
- **テストファイル**: `*.test.ts` または `*.spec.ts`

### モジュール構成
```typescript
// 各モジュールの標準構造
src/
├── index.ts              # エントリポイント
├── types.ts             # 型定義
├── config.ts            # 設定
├── handlers/            # リクエストハンドラー
├── services/            # ビジネスロジック
├── utils/               # ユーティリティ
└── __tests__/           # テスト
```

## スペック間連携パターン

### 1. API通信規約
```typescript
// 標準レスポンス形式
interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
  timestamp: string;
}
```

### 2. 認証・承認
- **内部API**: Bearer Token + HMAC署名
- **外部API**: 各サービス固有の認証方式
- **ユーザー識別**: Discord User ID基準

### 3. エラーハンドリング
```typescript
// 統一エラーハンドリング
class AimCoachError extends Error {
  constructor(
    public code: string,
    public message: string,
    public context?: Record<string, any>
  ) {
    super(message);
  }
}
```

## データ管理パターン

### データベース設計原則
- **正規化**: 第3正規形まで適用
- **インデックス**: クエリパフォーマンス最適化
- **外部キー**: 参照整合性保証

### キャッシュ戦略
- **KV利用**: 頻繁アクセスデータ（ユーザー設定、セッション）
- **TTL設定**: データ性質に応じた適切な有効期限
- **無効化**: データ更新時の適切なキャッシュクリア

### マイグレーション管理
- **バージョン管理**: SQLマイグレーションファイル
- **ロールバック**: 安全な巻き戻し機能
- **ゼロダウンタイム**: 無停止でのスキーマ変更

## テストパターン

### 単体テスト
```typescript
// Jest + TypeScript設定
describe('ModuleName', () => {
  beforeEach(() => {
    // セットアップ
  });

  it('should handle normal case', async () => {
    // テストケース
  });
});
```

### 統合テスト
- **API間連携**: モックを使った各スペック間通信テスト
- **データベース**: テスト用DB環境での実際のクエリテスト
- **外部API**: モック・スタブを活用したテスト

### E2Eテスト
- **Discord Bot**: 実際のDiscord環境でのシナリオテスト
- **ユーザーフロー**: 登録からスコア分析まで全体フロー

## Gitワークフローとブランチ管理

### ブランチ戦略

**メインブランチ**:
- **main**: 本番環境にデプロイされる安定版コード
- **develop**: 開発統合ブランチ、staging環境への自動デプロイ

**作業ブランチ**:
- **feature/[機能名]**: 新機能開発（例: `feature/analytics-engine`）
- **fix/[修正内容]**: バグ修正（例: `fix/gemini-api-timeout`）
- **docs/[文書名]**: ドキュメント更新（例: `docs/steering-git-workflow`）

### コミット粒度とメッセージ規約

**コミット粒度原則**:
- **論理的単位**: 1つの論理的変更を1コミットにまとめる
- **独立性**: 各コミットは単独でビルド・テストが通る状態
- **最小単位**: ファイル単位ではなく機能単位での分割
- **段階的実装**: 大きな機能は複数の段階的コミットに分割

**コミットメッセージ形式**:
```
[type]: [簡潔な変更内容]

[詳細な説明（必要に応じて）]
[Breaking Changes（必要に応じて）]
```

**コミットタイプ**:
- **feat**: 新機能追加
- **fix**: バグ修正
- **refactor**: リファクタリング
- **docs**: ドキュメント変更
- **test**: テスト追加・修正
- **chore**: ビルド・設定変更

**コミット例**:
```
feat: implement Gemini API connector for score analysis

- Add GeminiConnectorService with rate limiting
- Implement structured prompt generation for aim analysis
- Add comprehensive error handling and retry logic

Breaking Changes: None
```

### プルリクエスト運用

**作成基準**:
- **feature/fix完了時**: 機能実装またはバグ修正完了時
- **レビュー必須**: 全てのPRはコードレビューを経てマージ
- **テスト必須**: CI/CDパイプラインを通過したもののみマージ可能

**PRタイトル形式**:
- `[WIP] feature: ` - 作業中（レビュー不要）
- `feat: ` - レビュー準備完了
- `fix: ` - バグ修正
- `docs: ` - ドキュメント更新

**マージ戦略**:
- **Squash and merge**: feature/fixブランチの複数コミットを1つにまとめる
- **履歴保持**: develop → main は通常のマージコミット
- **削除**: マージ後は作業ブランチを削除

### スペック駆動開発との連携

**ブランチとスペックの対応**:
- `feature/analytics-engine` → `.kiro/specs/analytics-engine/`
- コミット時に該当スペックのタスク状況を更新
- タスク完了時は明確にコミットメッセージに記載

**コミット粒度の例**:
```
feat: implement data processor for score normalization

- Complete task 2.1: AimLab and Kovaaks score parsing
- Add validation for CSV/JSON format detection
- Implement unified score data structure

Tasks completed: 2.1
Requirements: 1.1, 4.1
```

## デプロイメントパターン

### 環境分離
- **development**: ローカル開発環境
- **staging**: 本番同等のテスト環境（develop自動デプロイ）
- **production**: 本番環境（mainからの手動デプロイ）

### CI/CDパイプライン
1. **コード変更** → GitHub push
2. **自動テスト** → 全テストスイート実行
3. **ビルド** → TypeScript→JavaScript変換
4. **デプロイ** → Cloudflare Workers更新
5. **ヘルスチェック** → 動作確認

### ロールバック戦略
- **即座復旧**: 前バージョンへの自動切り戻し
- **段階的デプロイ**: カナリアリリース対応
- **監視アラート**: 異常検知時の自動通知

## 依存関係管理

### パッケージ管理
- **npm/yarn**: 外部ライブラリ管理
- **バージョン固定**: 予期しない更新回避
- **セキュリティ更新**: 定期的な脆弱性チェック

### 内部モジュール
- **共通ライブラリ**: shared/ディレクトリで一元管理
- **型定義共有**: 各スペック間での型定義統一
- **循環依存回避**: 明確な依存関係方向