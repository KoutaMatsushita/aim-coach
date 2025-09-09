# Tasks: ingestion-exporter-v1 — Exporter (Go) Focus

Scope: 本タスク群は Exporter(Go) に限定。Workers/API/DB マイグレーションは別提出。各タスクは1–3ファイル/<=0.5日を目安。括弧内は対象ファイルと要件参照。

## T1. CLI サブコマンド骨組み
- 目的: `link/export/unlink/status` をサブコマンド化、共通フラグ導入。
- 作業:
  - `cmd/aimcoach-exporter/main.go` を cobra 互換の簡易実装に整理（標準flag継続でも可）
  - `internal/cli/cli.go`: サブコマンド登録とフラグ定義（api-endpoint, api-token, log-level）
- 完了条件: `--help` で各コマンド説明/フラグが表示。（R1）

## T2. 設定解決とログ
- 目的: CLI>ENV 優先で設定を解決、構造化ログ出力。
- 作業:
  - `internal/config/config.go`: ENV名/既定値を定義、Resolve() 実装
  - `internal/log/logger.go`: level制御/JSONログ/Windowsでもカラー無効
- 完了条件: `status` で実効設定を表示、log-level 変更が反映。（R1, NFR-Usability）

## T3. Token ストア（DPAPI）
- 目的: access/refresh を安全に保存/取得/削除。
- 作業:
  - `internal/auth/tokenstore/store.go`: インターフェース（Get/Set/Delete/Rotate/Expiry）
  - `internal/auth/tokenstore/windows_dpapi.go`: Windows DPAPI 実装（ユーザースコープ）
  - `internal/auth/tokenstore/store_test.go`: モックバックエンドでの単体テスト（保存/取得/削除/期限）
- 完了条件: `link` 後に保存、`status` に期限表示、`unlink` で削除。（R1, R5, Security）

## T4. Link コマンド
- 目的: 6桁コードで端末リンク、トークン保存。
- 作業:
  - `internal/auth/link.go`: `POST /link/complete {code, device}` 呼び出し
  - `cmd/link.go`: `link --code <XXXXXX> [--device <name>]`
  - `internal/auth/link_test.go`: 正常系/エラー時のハンドリング（HTTPモック）
- 完了条件: 2xxで access_token/refresh_token を保存、エラー表示（再試行ガイド）。（R5, Security）

## T5. Refresh フロー
- 目的: 送信前チェックと401時の自動更新。
- 作業:
  - `internal/auth/refresh.go`: 有効期限≤5日で `/token/refresh`、401時は1回だけrefresh→再送
  - `internal/send/http.go`: BeforeSend フックでトークン取得、失敗時のハンドリング
  - `internal/auth/refresh_test.go`: 期限閾値/401時の再送/失敗パスの単体テスト
- 完了条件: 有効期限が近い/401で自動更新、失敗時は明確なエラー。（Design §9.1, R5）

## T6. Digest 算出ユーティリティ
- 目的: 正規化レコードからカノニカルJSON→SHA-256。
- 作業:
  - `internal/model/digest.go`: 丸め(1e-6)/キーソート/NaN拒否/ハッシュ
  - 単体テスト `internal/model/digest_test.go`
- 完了条件: 同一入力→同一digest、丸め・順序の差異で不変。（R6）

## T7. Kovaaks CSV 正規化
- 目的: Stats.csv をセッション単位に集約しレコード化。
- 作業:
  - `internal/kovaaks/parser.go`: ファイル名解析(task/mode/taken_at)、行集約(shots/hits/accuracy, damage, efficiency)、raw_json 保持
  - `internal/kovaaks/parser_test.go`: サンプルCSVでテスト
- 完了条件: `--kovaaks-csv-dir` で複数ファイル→NormalizedRecord[]、digest付与。（R1, R3, R6）

## T8. AimLabs SQLite 取込（Klutch.bytes）
- 目的: 既知テーブルを走査し正規化。
- 作業:
  - `internal/aimlabs/sqlite_reader.go`: modernc.org/sqlite で接続、PRAGMA列挙→既知テーブル（CircleShotData, CircleTrackData, Composite, MetricsData 等）選択的に読み取り
  - `internal/aimlabs/normalize.go`: レコード→NormalizedRecord（timestamp/mode/map/metrics/score/raw_json）
  - `internal/aimlabs/sqlite_reader_test.go`: output-sample でテスト（存在テーブルのみ）
- 完了条件: `--aimlabs-sqlite` 指定でNormalizedRecord[]生成、digest付与。（R1, R3, R6）

## T9. 送信クライアント強化
- 目的: 冪等送信と再試行。
- 作業:
  - `internal/send/http.go`: `Idempotency-Key` ヘッダ対応、指数バックオフ(2s/4s/8s+jitter)、429/5xxで再試行
  - `internal/send/client_test.go`: リトライ制御のユニットテスト（モック）
- 完了条件: 2xxで成功、再試行ロジックが期待通り。（R2, R6, NFR-Perf）

## T10. Export 実行フロー
- 目的: 読取→正規化→バッチ送信 or dry-run。
- 作業:
  - `cmd/export.go`: since/max-records/source/idempotency-key 対応、プロバイダ別にレコード収集
  - `internal/export/run.go`: バッチング/統計ログ（accepted/rejected）
  - `internal/export/run_test.go`: dry-run時の出力整形/バッチング検証
- 完了条件: `export` で標準出力 or 送信、統計出力。（R1, R2, Usability）

## T11. Unlink/Status
- 目的: トークン無効化/状態確認。
- 作業:
  - `cmd/unlink.go`: `/unlink` 呼び出し（サーバ側APIが用意されるまでローカル削除のみでも可）
  - `cmd/status.go`: 有効期限、参照ディレクトリ、保管先を表示
  - `cmd/status_test.go`: 表示内容のスナップショットテスト（ゴールデンファイル）
- 完了条件: unlink で安全に削除、status 出力が見やすい。（R5, Security）

## T12. ビルド/配布
- 目的: Windows 単一バイナリ配布。
- 作業:
  - `Makefile` or `scripts/build.ps1`: `GOOS=windows GOARCH=amd64`、バージョン埋め込み(`-ldflags -X`)
  - 署名/ハッシュ出力（任意）
- 完了条件: バイナリ生成、`--version` 表示。（NFR-Usability）

## T13. Lint/CI
- 目的: 品質自動チェック。
- 作業:
  - `golangci.yml` 追加、`make lint`/`make test` 用意
  - mise エントリ（go 1.22.x）
- 完了条件: CIで lint/test が走る。（Structure: Tooling）

## T14. クロスプラットフォーム TokenStore（任意拡張）
- 目的: Windows以外でも安全に保存（Keychain/Secret Service/KWallet/file AES-KMS）。
- 作業:
  - `internal/auth/tokenstore/keyring.go`: 99designs/keyring バックエンド
  - `internal/auth/tokenstore/file.go`: AES-256-GCM + Argon2 パスフレーズ
  - `internal/auth/tokenstore/keyring_test.go` / `file_test.go`
- 完了条件: `--token-store auto|wincred|keychain|secret-service|file|kms` で切替、statusにストア種別表示。（Design §9.1）

## 依存関係・順序
- T1→T2/T3→T4/T5→T6→T7/T8→T9→T10→T11→T12/T13

## アウトオブスコープ（本タスクセット）
- Workers/API/DB: 別タスクセットで提出（/ingest/*, /link/complete, /token/refresh, D1 マイグレーション等）
