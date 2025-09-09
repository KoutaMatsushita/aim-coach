Aim Coach Exporter (Go)

概要
- Windows向けの単一バイナリCLI。AimLabsのSQLite(Klutch.bytes)とKovaaksのCSVを読み取り、正規化JSONに変換してAPI(`/ingest/*`)にPOST、またはドライランで標準出力に出力します。

ビルド（ローカル）
- 要件: Go 1.22+
- コマンド: `go build ./cmd/aimcoach-exporter`

実行例（ドライラン）
```
aimcoach-exporter \
  --aimlabs-sqlite "C:\\Users\\<User>\\AppData\\Local\\Aim Lab\\Saved\\SaveGames\\Klutch.bytes" \
  --kovaaks-csv-dir "C:\\path\\to\\KovaaKs\\Stats" \
  --dry-run
```

実行例（送信）
```
aimcoach-exporter \
  --aimlabs-sqlite "C:\\...\\Klutch.bytes" \
  --kovaaks-csv-dir "C:\\...\\KovaaKs" \
  --api-endpoint "https://example.workers.dev" \
  --api-token "<TOKEN>"
```

設定方法
- CLI優先、ENVフォールバック。
  - CLI: `--api-endpoint`, `--api-token`, `--aimlabs-sqlite`, `--aimlabs-csv-dir`(任意), `--kovaaks-csv-dir`, `--dry-run`
  - ENV: `API_ENDPOINT`, `API_TOKEN`, `AIMLABS_SQLITE_PATH`, `AIMLABS_CSV_DIR`, `KOVAAKS_CSV_DIR`

出力フォーマット（NormalizedRecord）
```
{
  "provider": "kovaaks" | "aimlabs",
  "task": "XYSmall" | "CircleShot" | ...,
  "mode": "Challenge" | "...",
  "map": "string | null",
  "taken_at": "RFC3339",
  "metrics": { "shotsTotal": 123, "hitsTotal": 100, ... },
  "score": 123.45,           // あれば
  "raw_json": { ... }        // 元データ断片
}
```

注意
- `modernc.org/sqlite` を使用してSQLiteを読み取ります（CGO不要）。
- 本リポでは依存解決は行っていません。ネットワーク許可後に `go mod tidy` を実行してください。

