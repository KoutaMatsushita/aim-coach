# Requirements Document: ingestion-exporter-v1

## Introduction

本MVPでは、AimLabs（ローカルSQLite: Klutch.bytes）およびKovaaks（CSV）のスコアデータを正規化してCloudflare WorkersのIngestion APIに送信し、Cloudflare D1へ保存する。Windows向けCLIエクスポーター（Go優先）でデータ収集を自動化し、Discordのリンクフローで発行されたトークンにより認可する。

## Alignment with Product Vision

- プロダクトの目的「FPSエイム練習の効率最大化」を支援する基盤機能。
- スコア収集アグリゲーションとパフォーマンス分析（将来）に必須となるデータ取り込みを確立。
- Tech Steering（Workers + D1、Mastraワークフロー、Discord 3秒ACK方針）に準拠。

## Requirements

### Requirement 1: Windowsエクスポーター（Go CLI）

User Story: FPS練習者として、手作業なしで練習結果が自動で取り込まれてほしい。そうすれば履歴が勝手に可視化・分析され、次アクション提案につながる。

Acceptance Criteria
1. WHEN ユーザーがCLIに初回起動時の`/link`コードを入力 THEN システム SHALL 長期APIトークンを取得・安全に保存（DPAPI/資格情報マネージャ）。
2. WHEN 実行時に`--aimlabs-sqlite`/ENVが与えられる THEN システム SHALL Klutch.bytes(SQLite)を直接読み取り、正規化JSONを生成。
3. IF `--kovaaks-csv-dir` が与えられる THEN システム SHALL KovaaksのStats CSV群を読み取り、セッション単位の正規化JSONを生成。
4. WHEN `--dry-run` 指定 THEN システム SHALL 送信せずに標準出力へJSON出力。
5. WHEN `--api-endpoint` と `--api-token` が設定 THEN システム SHALL `/ingest/aimlabs` および `/ingest/kovaaks` にHTTPS POSTし、2xxをもって成功とみなす。

CLI コマンド/オプション（定義）
- サブコマンド:
  - `link` … Discordで取得したコードを使って端末をリンク（トークン発行）
  - `export` … AimLabs/Kovaaks を読み取り、`--dry-run` または送信
  - `unlink` … 端末の長期トークンを失効（ローカルからも削除）
  - `status` … 認証状態と設定を表示（トークンの有効期限、参照パス）
- 共通オプション:
  - `--api-endpoint <url>` / `API_ENDPOINT`
  - `--api-token <token>` / `API_TOKEN`（`link`後は自動読込、手動指定で上書き可）
  - `--log-level <debug|info|warn|error>`（既定: info）
- `export` 専用:
  - `--aimlabs-sqlite <path>` / `AIMLABS_SQLITE_PATH`
  - `--aimlabs-csv-dir <dir>` / `AIMLABS_CSV_DIR`（開発用・任意）
  - `--kovaaks-csv-dir <dir>` / `KOVAAKS_CSV_DIR`
  - `--dry-run`（送信せず標準出力）
  - `--idempotency-key <uuid>`（任意。再送時のリクエスト冪等化）
  - `--source <string>` 端末識別メタ（例: `desktop-01`）
  - `--max-records <n>` 読み取り上限（既定: 無制限）
  - `--since <RFC3339>` 差分取り込みの開始時刻

### Requirement 2: Ingestion API（Workers）

User Story: サービス運営者として、外部エクスポーターから受け取ったスコアを安全・確実に保存したい。そうすれば分析・可視化・LLMコーチングの土台が安定する。

Acceptance Criteria
1. WHEN POST `/ingest/aimlabs` OR `/ingest/kovaaks` with Bearer THEN システム SHALL トークン検証を行い、正規化レコード配列を受理する。
2. WHEN バリデーション成功 THEN システム SHALL D1へ挿入し、`{accepted, rejected}` をJSONで返す（200）。
3. IF 重複（同一ユーザー・provider・task・taken_at・digest）が検出された場合 THEN システム SHALL 冪等的にスキップし、`rejected`へカウント。
4. WHEN エラー THEN システム SHALL 部分的に受理し、`errors[]`に要因を付与（バリデーションキー/行番号）。
5. IF リクエストヘッダ `Idempotency-Key` が再送で一致 THEN システム SHALL 24h以内は同一レスポンスを返す。

### Requirement 3: 正規化スキーマ

User Story: 開発者として、提供元に依存しない形でスコアを扱いたい。そうすれば将来の可視化/分析/LLMが一貫したI/Fで実装できる。

Acceptance Criteria
1. WHEN レコード生成 THEN システム SHALL 以下の構造に従う:
   - `provider: "aimlabs"|"kovaaks"`
   - `task: string`, `mode?: string`, `map?: string`
   - `taken_at: RFC3339(UTC)`
   - `metrics: object`（存在キーのみ、数値中心）
   - `score?: number`, `raw_json?: object`
2. WHEN D1保存 THEN システム SHALL 下記のハイブリッド構成で永続化する。
   - 統合: `scores(id, user_id, source_id, provider, task, mode, map, taken_at, metrics_json, score, raw_json, digest, created_at)`
   - 生データ: `aimlabs_raw(id, user_id, play_id, table_name, row_json, taken_at, created_at)` / `kovaaks_raw(id, user_id, file_name, row_json, taken_at, created_at)`
   - インデックス: `(user_id, taken_at)`, `(provider, task, taken_at)`, `UNIQUE(digest)`
   - `digest`: 正規化レコードのカノニカルJSONを `SHA-256` で算出（小数は桁丸め、キー順序固定）。

### Requirement 4: Discordリンク/3秒ACK関連（関連要件）

User Story: ユーザーとして、Discordで簡単にリンク設定したい。そうすればエクスポーターが安全にアップロードできる。

Acceptance Criteria
1. WHEN `/link` コマンド THEN システム SHALL 3秒以内に`type:5`で即ACK（必要に応じ`flags:64`）。
2. THEN システム SHALL `waitUntil`で短命トークン発行→長期APIトークンへ交換し、UI上で案内（元メッセージ編集）。

### Requirement 5: Link 仕様（詳細）

User Story: 初回セットアップ時にブラウザを開かず、DiscordとCLIだけで端末を安全にリンクしたい。

Acceptance Criteria
1. WHEN ユーザーが `/link` を実行 THEN システム SHALL 一意な6桁コードと有効期限（例: 10分）を生成し、Discordにエフェメラル表示。
2. WHEN CLIが `aimcoach-exporter link --code <XXXXXX>` を呼ぶ THEN システム SHALL `POST /link/complete` により `access_token(30d)` と `refresh_token(90d)` を返却。
3. THEN CLI SHALL Windows DPAPI（ユーザースコープ）でトークンを暗号化保存し、以降の `export` では自動付与。
4. WHEN `refresh_token` 期限切れが近い THEN CLI SHALL 送信前に自動更新（リトライ/バックオフ）。
5. WHEN `unlink` THEN システム/CLI SHALL 即時にトークンを失効し、ローカル保存も削除。
6. Security: コードは一回限り・短命・推測困難。5回連続失敗でコードを失効。
7. Non-goals(MVP): Discord OAuth（ブラウザ起動）は採用しない。必要時に拡張可能とする。

### Requirement 6: 冪等性（送信側/受信側）

User Story: ネットワーク断や再実行時でも重複保存を起こさず、安全に再送したい。

Acceptance Criteria
1. 送信側（CLI）: レコードごとに `digest` を算出し、ボディに含める。任意で `--idempotency-key` を指定した場合、同一リクエスト単位の冪等化も行う。
2. 受信側（Workers）: `scores.digest` の一意制約で重複を拒否。レスポンスの `rejected` に計上し、既存IDを返す。
3. 受信側: ヘッダ `Idempotency-Key` が一致する24h以内の再送には最初のレスポンスを返す（ステータス/ボディ完全一致）。
4. ログ/監査: `request_id` と `idempotency_key`、`digest` を関連付けて記録。

## Non-Functional Requirements

### Code Architecture and Modularity
- エクスポーター: 入力(adapter)・正規化(core)・送信(transport)を分離。
- Workers: ルーティング/バリデーション/永続化層を分離。Mastraワークフロー適用箇所をモジュール化。

### Performance
- エクスポーター1回の実行で1,000レコード規模を処理可能（< 10秒目標・ローカルIO依存）。
- Ingestionは1リクエストあたり最大レコード数を10,000未満に制限、500KB/レコード上限。

### Security
- Bearerトークン必須（Discordリンクで発行）。
- HTTPSのみ、署名/トークンをログに出さない。

### Reliability
- 冪等性（digestベースの重複排除）。
- 部分受理（成功分は保存、失敗分はerrorsへ）。

### Usability
- CLI/ENV両対応。`--help`整備。失敗時は原因を日本語で表示。
