# Technology Stack

## Project Type
Agentベースのコーチングサービス。MVPは「Cloudflare Workers（API/Discordインタラクション）+ Cloudflare D1（DB）+ Windowsローカルのスコアエクスポータ」で構成。Webダッシュボードは後続。

## Core Technologies

### Primary Language(s)
- Language: TypeScript（ESM）
- Runtime: Cloudflare Workers（本番） / Node.js >= 20.9（ローカル開発・ツール）
- Dev CLIs: `mastra`（主: `mastra deploy` + CloudflareDeployer）、`wrangler`（ローカル検証/D1操作）、`npm`

### Key Dependencies/Libraries
- AI/LLM: AI SDK 経由でプロバイダ抽象（初期は `@ai-sdk/google` を想定、Workers互換を優先。互換性が不足する場合はHTTPクライアント実装に切替）
- Validation: zod
- Logging: Pino（Workers互換の軽量ロガー設定）
- Discord Interactions: 署名検証とコマンドハンドラ（例: `discord-interactions` or 署名検証の自前実装）

### Application Architecture
- Monolith on Workers + 周辺ツール
  - Ingestion API（Workers）: JSON受け付け（`/ingest/aimlabs`, `/ingest/kovaaks`）
  - Windows Exporters: AimLabs/KovaaksのローカルデータをJSONへ変換しHTTPS送信（初期はCLIアプリ）。
    - 推奨実装: Go 1.22（単一バイナリ/ランタイム不要、Windows配布容易）
    - 代替候補: Deno/Bun（TSで高速実装可能、SQLite/CSV対応の検証後に選定）、.NET 8（実績豊富）
  - Analytics: 指標分解の統計処理（Workers内で軽量に実施）
  - LLM Orchestrator: スコア履歴＋会話ログを入力に“次アクション”生成（Edge互換LLMクライアント）
  - Conversation Agent: Discord Slash Commands + Follow-up Webhooks（Gateway接続は行わない）
  - Storage: Cloudflare D1（ドメインデータ＋会話ログ要約＋アドバイス）
  - Queues/Crons（将来）: 定期レビューやリマインド配送

### Data Storage
- Primary storage: Cloudflare D1
- Schemas（初期案）
  - users(id, discord_user_id, created_at)
  - sources(id, kind['aimlabs'|'kovaaks'|'manual'], meta, user_id, created_at)
  - scores(id, user_id, source_id, task, taken_at, raw_json, metrics_json)
  - conversations(id, user_id, channel, role, content, ts)
  - advice(id, user_id, session_id, summary, drills_json, created_at)
- Data formats: IngestionはJSON（Windowsエクスポータが生成）
- Caching: 当面不要（必要に応じて KV/内蔵キャッシュ）

### External Integrations
- Discord Interactions（Slash Commands / Component / Follow-up Webhooks）
  - インタラクション署名検証（Ed25519）
  - エンドポイントURL: Workers で公開
- LLM Provider: AI SDK 経由（Edge互換）。非互換が判明した場合はHTTPベースへ切替
- AimLabs: ローカルSQLite（公式クライアントのローカルDB）からエクスポータが抽出→JSON化→Workersの`/ingest/aimlabs`へPOST（ToS順守）
- Kovaaks: CSVをエクスポータでJSON化→`/ingest/kovaaks`へPOST（Windowsから送信）

#### Windows Exporter（詳細）
- 言語/ランタイム: Go 1.22（単一バイナリ）。Deno/Bun/.NETは代替案として検討可
- 依存（Go想定）: `modernc.org/sqlite`（AimLabs SQLite読み取り・CGO不要）, 標準 `encoding/csv`（Kovaaks CSV読み取り）
- 機能:
  - スケジュール収集（タスクスケジューラに登録）と手動ワンショット実行
  - データ変換（正規化/バリデーション）→ JSON 生成
  - HTTPS POST（署名付き）で Workers の `/ingest/*` へ送信（再試行/バックオフ）
  - 単一実行形式: ユーザー配布容易性を重視（必須）

#### Discord 連携（リンクフロー）
- ユーザーは Discord のスラッシュコマンド（例: `/link`）でペアリングコードを取得
- Windows Exporter は初回起動時にペアリングコードを入力→ Workers 経由で短命トークンを交換→ 長期 API トークンを安全に保存（DPAPI/ユーザースコープ）
- 以降の送信は長期トークンで認可（D1 の `users` と `sources` に紐づく）

#### Discord スラッシュコマンドの3秒対応
- 基本方針: 署名検証後すぐにACKし、重い処理はバックグラウンドで実行して「後から編集/フォローアップ」する。
- 具体:
  - PINGは `type: 1`（PONG）で即応答。
  - コマンドは `type: 5`（DEFERRED_*）で即ACK。必要に応じて `flags: 64`（ephemeral）。
  - Workersでは `ctx.waitUntil(...)` で非同期処理を継続。
  - 完了時は Webhook 経由で「元メッセージ編集」`PATCH /webhooks/{app_id}/{token}/messages/@original` もしくは「フォローアップ」`POST /webhooks/{app_id}/{token}`。
  - インタラクショントークンは短時間のみ有効なため、長時間処理はキュー（Queues）や別経路（Bot Token→DM）に切替可能性を考慮。
- 運用/安全:
  - 冪等性: `interaction.id` を D1 に記録し二重実行を防止。
  - 署名検証: Ed25519（`X-Signature-Ed25519`/`X-Signature-Timestamp`）。
  - タイムアウト時のUX: 先にACKで「処理中…」を表示し、失敗時は原因と再試行方法を編集で通知。
  - レート制御: フォローアップ送信にバックオフを実装。

```ts
// Workers イメージ（抜粋）
if (i.type === 1) return json({ type: 1 })
if (i.type === 2) {
  ctx.waitUntil(handleCommand(i, env))
  return json({ type: 5, data: { flags: 64, content: '処理中…' } })
}
```

#### Mastra ワークフロー（インポートのインターフェイス）
- スコアインポート処理は Mastra のワークフローとして記述（parse → validate → enrich → persist → notify）
- Workers 側の `/ingest/*` は「ワークフローアダプタ」。受信したペイロードを `mastra` ランタイム（Edge互換）へ渡して実行
- ドキュメント/運用連携に `@mastra/mcp-docs-server` を使用（ワークフロー定義の可視化/操作）
- 主要タスク例:
  - `ingest.aimlabs`: SQLite由来JSONの検証/変換/保存
  - `ingest.kovaaks`: CSV由来JSONの検証/変換/保存
  - `post.ingest.summarize`: 直近セッションの要約/アドバイス起動

#### Ingestion 入力ソース（ファイル）
- AimLabs（必須）: SQLite DB ファイル `Klutch.bytes`（ユーザー環境の既存DB。拡張子は独自だが中身はSQLite）
  - 設定キー: `AIMLABS_SQLITE_PATH`（例: `C:\\Users\\<User>\\AppData\\Local\\Aim Lab\\Saved\\SaveGames\\Klutch.bytes`）
  - CSVは開発/デバッグ用の代替（任意）: `AIMLABS_CSV_DIR`（今回の `./output-sample/aimlabs` 相当）
- Kovaaks（必須）: 出力CSVディレクトリ
  - 設定キー: `KOVAAKS_CSV_DIR`（例: `C:\\Users\\<User>\\Documents\\KovaaK's\u00AE\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f\u200e\u200f` 配下）

##### Exporter 設定方法（ENV と CLI の両対応）
- ENV 変数: `API_ENDPOINT`, `API_TOKEN`, `AIMLABS_SQLITE_PATH`, `AIMLABS_CSV_DIR`, `KOVAAKS_CSV_DIR`
- CLI オプション例:
  - `--api-endpoint <url>`
  - `--api-token <token>`
  - `--aimlabs-sqlite <path>`
  - `--aimlabs-csv-dir <dir>`（任意）
  - `--kovaaks-csv-dir <dir>`
- プリオリティ: CLI > ENV（両方未指定なら対話プロンプト or 設定ファイルにフォールバック）。

#### AimLabs（SQLite）→ 正規化
- 取得方針: DBスキーマを動的に列挙（`PRAGMA table_info`）し、既知テーブル（例: AudioSpatialData, CircleShotData, CircleTrackData, DecisionShotData, DodgeShotData, FreeTrackData, Composite, MetricsData, Coaching* 等）を走査。
- 同一 `PlayID`/`timestamp` 単位で集約し、以下の正規化レコードに変換。
- 正規化レコード（例）:
  - `provider`: `"aimlabs"`
  - `task`: モードやデータセットに応じた識別子（例: `CircleShot`, `CircleTrack` など）
  - `mode`: CSV/SQLite中の `mode`
  - `map`: CSV/SQLite中の `map`
  - `taken_at`: `timestamp`（UTC変換）
  - `metrics`: 可変フィールド（例: `accTotal`, `KPS`, `SPK`, `killTotal`, `shotsTotal`, `hitsTotal`, `missesTotal`, `targetsTotal`, `OTR`, `avgTimeOn/Off` など、存在するもののみ）
  - `score`: `score` が存在する場合のみ設定
  - `raw_json`: 元行/行群をJSONで保存（監査/将来解析用）

#### Kovaaks（CSV）→ 正規化
- ファイル名から `task` と `taken_at` を抽出（例: `XYSmall - Challenge - 2025.08.04-23.44.33 Stats.csv`）
- ヘッダ `Kill #, Timestamp, Bot, Weapon, TTK, Shots, Hits, Accuracy, ...` を行単位で読み取り、セッション集約を実施。
- 正規化レコード（例）:
  - `provider`: `"kovaaks"`
  - `task`: 例 `XYSmall`
  - `mode`: 例 `Challenge`
  - `taken_at`: 例 `2025-08-04T23:44:33Z`
  - `metrics`: `shotsTotal`, `hitsTotal`, `accuracy`（`Hits/Shots` の加重平均）、`damageDone`, `damagePossible`, `efficiency` 等
  - `raw_json`: 1セッション分の行配列

### Monitoring & Dashboard Technologies
- ログ: Workersログ + 可観測性（将来: Logpush 連携）
- 可視化: 後続のWebダッシュボードでスコア推移/弱点レーダー/次アクションカード

## Development Environment

### Build & Development Tools
- Build/Run: `mastra dev`（ワークフロー実行/可視化）, `wrangler dev`（Workers）, `wrangler d1`（スキーマ/クエリ管理）, npm scripts
- Exporter: Go（`go build`）/ Deno（`deno compile`）/ Bun（`bun build`）のいずれかで単一バイナリ化
- Development workflow: Mastra中心（ワークフロー駆動） + wrangler によるエッジ検証
- Package Management: npm

### Code Quality Tools
- Static Analysis: TypeScript（strict）
- Formatting: Prettier
- Testing: Vitest（API/ユーティリティ）, Miniflare（Workersローカル）
- Docs: Steering/Specドキュメント

### Version Control & Collaboration
- VCS: Git
- Branching: trunk-based or GitHub Flow
- Review: PRレビュー、ステアリング/スペック準拠

### Dashboard Development
- 後続。Workers + Pages/SSRなどは次期で定義。

## Deployment & Distribution
- Target: Cloudflare Workers（API/Discordインタラクション） + Cloudflare D1（DB）
- Distribution: `mastra deploy`（CloudflareDeployer を使用）を優先。`wrangler` はローカル検証/緊急対応で併用
- Installation Requirements: Discordアプリ設定（インタラクションエンドポイントURL/権限）、D1バインディング、Mastra用環境変数
- Update Mechanism: Mastra のリリースパイプライン + D1マイグレーション（`wrangler d1 migrations`）

### セキュリティ/設定（Exporter）
- `API_ENDPOINT`, `API_TOKEN`（Workers 側で発行される長期トークン）
- `AIMLABS_SQLITE_PATH`, `AIMLABS_CSV_DIR?`, `KOVAAKS_CSV_DIR`
- Windows 資格情報マネージャ/DPAPI でトークンを暗号化保存

## Technical Requirements & Constraints

### Performance Requirements
- Discord応答: 初期ACK 3秒以内、フォローアップで最終応答
- Ingestion: JSON 1万行/分程度（バルク/トランザクション。D1制限内で設計）
- LLM: ストリーミング応答に対応（ユーザー体感向上）

### Compatibility Requirements
- Platform: Cloudflare Workers（主）、Windows（エクスポータ実行環境）
- Dependencies: discord-interactions互換、Edge対応のLLMクライアント
- Standards: Discord Rate Limit/署名検証、HTTP/JSON

### Security & Compliance
- Secrets: Workers Secrets（Bot Token/LLM API Key）
- Data: 会話/スコアのプライバシー（同意・削除/匿名化API）
- LLM安全策: ガードレール/根拠提示/再生成
- ToS順守: AimLabs/Kovaaksのデータ取り扱いは利用規約/著作権・逆コンパイル禁止に準拠

### Scalability & Reliability
- 期待負荷: 〜数百DAU（初期）。Workersで自動スケール
- 信頼性: Discordリトライ/レート制御、LLM障害フォールバック
- 将来: Durable Objects/KV/Queuesで状態やジョブを拡張

## Technical Decisions & Rationale

### Decision Log
1. Cloudflare D1を主DBに採用: サーバレスで運用容易。グローバル配信に適合。代替: LibSQL/Turso, Postgres。
2. DiscordはGateway接続ではなくInteractionsを採用: Workersで安定動作（長時間WS不要）。代替: discord.js Gateway（別ホストが必要）。
3. AimLabs/Kovaaks取り込み: WindowsエクスポータでローカルDB/CSV→JSON化→HTTPS送信。ToS順守。将来公式API/ツール連携を検討。
4. LLMクライアント: Edge互換を優先。非互換時はHTTP直呼出 or 別ランタイムのフォールバックを用意。

## Known Limitations
- Interactions中心のためDM/常時リッスン系は限定（必要なら別ランタイムでGatewayを補完）
- D1のクエリ/同時実行制約に注意（バルク挿入やインデックス設計が必要）
- Windowsエクスポータの配布/アップデート経路が必要（自動更新は後続）
- モデルコスト/レイテンシ最適化（プロンプト圧縮/要約/キャッシュ）が必要
