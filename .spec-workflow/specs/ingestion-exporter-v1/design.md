# Design Document: ingestion-exporter-v1

## 1. Scope & Goals
- 対象: AimLabs/Kovaaks スコア取り込みのMVP。
- 目的: Windows CLI Exporter で収集→Workers API (/ingest/*) で受信→D1 に正規化保存。Requirements版の要件・NFRを満たす詳細設計を定義。
- 非対象: ダッシュボード、LLMコーチング本体、Discord OAuth（将来拡張）。

## 2. Architecture Overview
- コンポーネント
  - Exporter(Go, Windows): `link/export/unlink/status`。DPAPIでトークン保存。AimLabs(SQLite) / Kovaaks(CSV) を正規化してPOST。
  - Workers(API): `POST /ingest/aimlabs`, `POST /ingest/kovaaks`。zod検証→D1保存→部分受理応答。`Idempotency-Key` 対応。
  - Workers(Discord): `/discord/interactions` 署名検証→3秒ACK(type:5)→waitUntilでトークン発行/編集。
  - D1: `scores`, `aimlabs_raw`, `kovaaks_raw`, `users`, `sources`。
  - Mastra(任意): `ingest.aimlabs`, `ingest.kovaaks`, `post.ingest.summarize` のワークフローをWorkers内から呼び出すアダプタ。

- データフロー（要約）
  1) Link: Discord `/link` → Workers が6桁コード発行 → CLI `link --code` → `access_token/refresh_token` 返却 → DPAPI保存。
  2) Export: CLI が AimLabs/Kovaaks から読み取り→正規化→`digest`算出→`/ingest/*` へPOST(Idempotency-Key任意)→結果表示。
  3) Persist: Workers は検証→rawテーブルへ保存（任意）→正規化`scores`をUPSERT（`UNIQUE(digest)`）。

## 3. Data Model (D1/SQLite)
- 共通注意: TEXTはUTF-8、タイムスタンプはUTC RFC3339文字列、JSONカラムはTEXT(JSON文字列)で保存。

### 3.1 users
- `id` TEXT PRIMARY KEY (ULID)
- `discord_user_id` TEXT UNIQUE NOT NULL
- `created_at` TEXT DEFAULT (CURRENT_TIMESTAMP)

### 3.2 sources
- `id` TEXT PRIMARY KEY (ULID)
- `user_id` TEXT NOT NULL REFERENCES users(id)
- `kind` TEXT CHECK(kind IN ('aimlabs','kovaaks','manual'))
- `meta` TEXT  -- JSON文字列（ホスト名、DPI等）
- `created_at` TEXT DEFAULT (CURRENT_TIMESTAMP)
- INDEX(user_id)

### 3.3 scores (正規化/横断分析用)
- `id` TEXT PRIMARY KEY (ULID)
- `user_id` TEXT NOT NULL REFERENCES users(id)
- `source_id` TEXT REFERENCES sources(id)
- `provider` TEXT CHECK(provider IN ('aimlabs','kovaaks')) NOT NULL
- `task` TEXT NOT NULL
- `mode` TEXT
- `map` TEXT
- `taken_at` TEXT NOT NULL  -- RFC3339 UTC
- `metrics_json` TEXT NOT NULL  -- {key:number,...}
- `score` REAL
- `raw_json` TEXT  -- 少量の生断片（監査/可視化向け）
- `digest` TEXT NOT NULL UNIQUE  -- 冪等用ハッシュ
- `created_at` TEXT DEFAULT (CURRENT_TIMESTAMP)
- INDEX(user_id, taken_at)
- INDEX(provider, task, taken_at)

### 3.4 aimlabs_raw / kovaaks_raw (再処理/検証用)
- aimlabs_raw:
  - `id` TEXT PRIMARY KEY, `user_id` TEXT, `play_id` TEXT, `table_name` TEXT,
  - `row_json` TEXT, `taken_at` TEXT, `created_at` TEXT DEFAULT CURRENT_TIMESTAMP
  - INDEX(user_id, taken_at), INDEX(table_name)
- kovaaks_raw:
  - `id` TEXT PRIMARY KEY, `user_id` TEXT, `file_name` TEXT,
  - `row_json` TEXT, `taken_at` TEXT, `created_at` TEXT DEFAULT CURRENT_TIMESTAMP
  - INDEX(user_id, taken_at)

## 4. Canonicalization & Digest
- 目的: 重複排除/冪等性。
- 対象フィールド: `provider, user_id, task, mode?, map?, taken_at, metrics`（小数は小数第6位で丸め）、`score?`。
- 手順:
  1. `metrics` のキーをソート、数値は `round(x, 1e-6)`、NaN/Infは拒否。
  2. オブジェクトをキー順固定のカノニカルJSONへ変換（スペースなし）。
  3. `SHA-256(hex)` を `digest` として設定。

## 5. Exporter (Go) Design
- 構成: `internal/{config, model, kovaaks, aimlabs, send}`、`cmd/aimcoach-exporter`。
- 設定解決: CLI > ENV。`status` で実効設定を表示。
- Link: `link --code <XXXXXX>` → `POST /link/complete`。成功時に `access_token(30d)` と `refresh_token(90d)` を受領しDPAPI保存。`unlink` は即失効。
 - トークン更新: 送信前に `access_token` の有効期限を確認し、残り≤5日で `POST /token/refresh` を実行。`401` 受信時は一度だけ自動リフレッシュ→再送。
- 読み取り:
  - AimLabs: modernc.org/sqlite で Klutch.bytes を開く。`PRAGMA table_info`で列挙→既知テーブルを走査→正規化。
  - Kovaaks: `* Stats.csv` を読み込み、ファイル名から `task/mode/taken_at` 抽出→行集約→正規化。
- 送信: `POST /ingest/<provider>`。ヘッダ `Authorization: Bearer`, `Idempotency-Key?`。`429/5xx` は指数バックオフで最大3回。
- 出力: `--dry-run` でJSONをstdoutへ（1行/レコード or pretty）。

## 6. Workers (API) Design
- ルーティング: Hono相当の軽量ルーター、`fetch` エントリから `/ingest/*` と `/discord/interactions` をディスパッチ。
- バリデーション: zodでNormalizedRecord配列を検証（数値・時刻・必須キー、metricsの最大キー数制限）。
- 永続化:
  - rawへの保存（任意/デバッグ向け）: 入力を一定粒度で `*_raw` に保存。
  - scores: `digest` 一意制約でUPSERT相当。既存ならrejectedとしてカウント。
- 応答: `{accepted: number, rejected: number, errors?: Array<{index:number, code:string, message:string}>}` を返す。
- Idempotency-Key: 24h以内は最初のレスポンスを返却（KV or D1でキャッシュ）。

## 7. Discord Interactions Design
- 署名検証: `verifyKey`。PONG(type:1)即返し。
- Slash: APPLICATION_COMMAND(type:2)は `type:5` で即ACK (flags:64既定)。`waitUntil` で短命コード発行→`/link/complete`を準備→元メッセージ編集。
- エラー時: 編集でガイダンス（再実行/サポート連絡）。

## 8. API Spec
### 8.1 POST /ingest/aimlabs | /ingest/kovaaks
- Headers: `Authorization: Bearer <token>`, `Content-Type: application/json`, `Idempotency-Key?`
- Body: `NormalizedRecord[]`
- 200 OK: `{accepted, rejected}` / `errors?`
- 400: バリデーション失敗、409: 重複(digest)

### 8.2 POST /link/complete
- Body: `{ code: string, device: { name?: string } }`
- 200 OK: `{ access_token: string, refresh_token: string, expires_in: number }`

### 8.3 POST /token/refresh
- Headers: `Authorization: Bearer <refresh_token>`
- Body: `{}`（不要）
- 200 OK: `{ access_token: string, refresh_token?: string, expires_in: number }`
  - セキュリティ上、毎回 `refresh_token` をローテーションする実装も可（設定で切替）。
 - 401: refresh が失効。ユーザーに `link` 再実行を案内。

## 9. Security
- 通信: HTTPS必須。
- 認可: Bearer（ユーザー/ソース紐付け）。トークンはRotate可能、`unlink`で即無効。
- 保存: DPAPI（ユーザースコープ）。ログにはトークン値を出さない。
- 署名: Discord Ed25519検証。

### 9.1 Token Lifecycle & Refresh
- 3トークン状態:
  - Access Token: 有効期限≈30日。送信時に付与。
  - Refresh Token: 有効期限≈90日。`/token/refresh` の認可に使用（Bearer）。
  - Pairing Code: `/link` で一時的に発行（10分）。一回限り。
- クライアント動作（Exporter）:
 1) 送信直前に `access_token.expires_at` を確認し、残り≤5日で `/token/refresh` 実行。
 2) `/ingest/*` が `401` の場合、1回だけ `refresh`→再送。2回目以降の401は失敗として案内。
 3) 成功時は新しい `access_token`（+必要に応じ `refresh_token`）をDPAPIに保存。旧 `refresh_token` は破棄。
 4) `unlink` はサーバ・ローカル両方で即無効化（rotate防止）。
 - サーバ動作:
  - Refresh Token にはローテーションID/世代を持たせ、盗用対策としてトリガー時に前世代を無効化。
  - 失効/取り消しはユーザー単位 or デバイス（source）単位で管理。

## 10. Performance & Limits
- 1リクエスト最大レコード: 10,000未満。1レコード上限: 500KB。
- Workersは `ctx.waitUntil` でDB挿入をバックグラウンド化可能（応答を早めるため）。
- バックオフ: 2s/4s/8s（jitterあり）。

## 11. Observability
- ログ項目: `request_id`, `user_id`, `source_id`, `provider`, `accepted`, `rejected`, `idempotency_key`, `digest[0:8]`。
- エラー分類: バリデーション/DB/権限/リトライ超過。

## 12. Migrations
- 0001_init.sql: users, sources, scores, aimlabs_raw, kovaaks_raw
- 0002_indexes.sql: 各テーブルのINDEXと `scores.digest UNIQUE`

## 13. Rollout & Config
- wrangler.toml: D1バインディング/Secrets。
- mise: Node>=20.9, Go 1.22.x。
- Biome(Workers)/golangci-lint(Exporter)。

## 14. Test Plan
- Exporter: 正規化/ファイル名パース/digest算出の単体テスト、サンプルCSV/SQLiteでE2E（dry-run）。
- Workers: バリデーション/部分受理/Idempotency-Keyの統合テスト（Miniflare）。
- 回帰: digest一意制約の衝突テスト。

## 15. Risks & Mitigations
- SQLiteスキーマ差異: 実環境差分はPRAGMA列挙+存在チェックで回避。
- 大量送信: 分割送信+レート制御。
- トークン漏洩: DPAPI+token rotate+unlink導線。

## 16. Future Work
- Discord OAuth同時提供、Cloudflare Queuesで非同期化、追加プロバイダ対応、ダッシュボード可視化。
