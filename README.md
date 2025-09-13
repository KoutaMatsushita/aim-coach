# AimCoach - AI-Powered Aim Training Coach

FPSプレイヤーのエイム技術向上を支援するAI駆動のコーチングプラットフォーム。AimLabとKovaaaksのスコアデータを分析し、Gemini APIベースの対話型インターフェースを通じて個別ユーザーに最適化されたエイムトレーニング指導を提供します。

## 🎯 プロジェクト状況

**✅ 完全リアーキテクチャ完了**

既存の実装を全て破棄し、Cloudflare Workers環境に最適化された新しいマイクロサービス型アーキテクチャで再構築。ステアリング設定完了により、一貫性のある開発基盤を確立。

**現在のフェーズ**: 要件定義準備完了 🚀

## 🏗️ アーキテクチャ

### マイクロサービス構成
プロジェクトは以下の4つの独立スペックに分割：

| スペック | 責務 | 技術スタック | 優先度 |
|---------|------|-------------|-------|
| **discord-bot-interface** | Discord Bot UI/UX | Cloudflare Workers + TypeScript | 🔥 **最優先** |
| **conversation-engine** | LLM対話処理 | Gemini API + Durable Objects | ⭐ 高 |
| **analytics-engine** | スコア分析・提案 | D1 Database + Gemini API | ⭐ 高 |
| **data-collector** | ローカルデータ収集 | Windows + API同期 | 🔺 中 |

### 技術スタック
```
🌐 Cloudflare Workers Ecosystem
├── 💾 Cloudflare D1 (SQLite)     # メインデータベース
├── ⚡ Cloudflare KV              # 高速キャッシュ
├── 🏠 Durable Objects            # セッション管理
└── 🔒 Workers Secrets            # API キー管理

🤖 AI & External APIs
├── 🧠 Gemini API                 # LLM対話・分析
└── 💬 Discord API                # Bot インターフェース

🛠️ Development Stack
├── TypeScript/JavaScript         # 型安全開発
├── Wrangler CLI                  # CF Workers 管理
└── Jest                          # テストフレームワーク
```

## 📋 開発フロー

### Kiro Spec-Driven Development
- **Steering設定**: ✅ 完了 ([product](.kiro/steering/product.md), [tech](.kiro/steering/tech.md), [structure](.kiro/steering/structure.md))
- **スペック管理**: [Kiro方式](./CLAUDE.md#project-context)による段階的開発
- **品質保証**: 要件 → 設計 → タスク → 実装の4段階承認

### 開発進捗
- [x] **Phase 0**: プロジェクトリセット・クリーンアップ
- [x] **Phase 1**: ステアリング設定（技術方針・構造定義）
- [x] **Phase 2**: 4スペック初期化
- [ ] **Phase 3**: 要件定義（`discord-bot-interface`から開始）
- [ ] **Phase 4**: 設計・実装（マイクロサービス順次開発）

## 🚀 次のアクション

### 1. 要件定義開始
```bash
# Discord Bot Interface から開始（最優先）
/kiro:spec-requirements discord-bot-interface
```

### 2. 推奨開発順序
1. 🔥 **discord-bot-interface** ← **ここから開始**
2. ⭐ **conversation-engine**
3. ⭐ **analytics-engine**
4. 🔺 **data-collector**

### 3. 並行開発可能性
- `conversation-engine` と `analytics-engine` は並行開発可能
- `data-collector` は他3つの完成後に実装推奨

## 📖 ドキュメント

| ファイル | 内容 |
|---------|------|
| [CLAUDE.md](./CLAUDE.md) | 開発フロー・コマンド詳細 |
| [.kiro/steering/](.kiro/steering/) | プロジェクト全体方針 |
| [.kiro/specs/](.kiro/specs/) | 各機能スペック |

## 🎯 品質目標

- **パフォーマンス**: Discord応答 < 3秒、分析処理 < 30秒
- **可用性**: 99.5%システム稼働率
- **ユーザー体験**: 自然な日本語対話、初心者配慮
- **セキュリティ**: プライバシー保護、API認証徹底

---

**Ready for Requirements Phase** 🚀 - 要件定義コマンド実行で開発開始