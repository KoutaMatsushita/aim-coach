# AimCoach - Discord Bot for Aim Training Analysis

AimCoachは、AimLabとKovaaaksのスコアデータを分析し、LLMベースの対話型インターフェースを通じて個別ユーザーに最適化されたエイムトレーニング指導を提供するCloudflareベースのDiscord Botです。

## プロジェクト状況

**⚠️ 注意: このプロジェクトは完全にゼロからリスタートしました**

既存の実装を全て破棄し、Cloudflare Workers環境に最適化された新しいアーキテクチャで再構築中です。

## アーキテクチャ

### 機能分割アプローチ
プロジェクトは以下の独立した機能に分割されています：

1. **Discord Bot Interface** - メインのDiscord Bot（優先実装）
2. **Analytics Engine** - スコア分析エンジン
3. **Data Collector** - スコア収集システム
4. **Conversation Engine** - LLM対話エンジン

### 技術スタック
- **Runtime**: Cloudflare Workers
- **Language**: TypeScript/JavaScript
- **Storage**: Cloudflare D1 (SQLite), KV Storage
- **Session Management**: Durable Objects
- **AI Integration**: OpenAI API
- **Interface**: Discord Bot API

## 開発フロー

### Kiro Spec-Driven Development
このプロジェクトは[Kiro方式のSpec-Driven Development](./CLAUDE.md#project-context)を使用しています。

### 現在のフェーズ
- [x] 既存コード削除とリポジトリリセット
- [ ] 機能分割されたスペック作成
- [ ] Discord Bot Interface の設計と実装
- [ ] その他機能の段階的実装

## 次のステップ

1. 機能ごとのスペック作成
2. Discord Bot Interface から実装開始
3. 段階的な機能追加

詳細な開発手順は [CLAUDE.md](./CLAUDE.md) を参照してください。