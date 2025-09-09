# Project Structure: aim-coach

## Directory Organization

```
project-root/
├── workers/                          # Cloudflare Workers（API/Discord）
│   ├── src/
│   │   ├── index.ts                  # fetchエントリ（Hono/Router）
│   │   ├── routes/
│   │   │   ├── ingest/
│   │   │   │   ├── aimlabs.ts       # POST /ingest/aimlabs（バリデーション→Mastraワークフロー）
│   │   │   │   └── kovaaks.ts       # POST /ingest/kovaaks
│   │   │   └── discord.ts           # POST /discord/interactions（3秒ACK→waitUntil）
│   │   ├── lib/
│   │   │   ├── validation/          # zodスキーマ（NormalizedRecord など）
│   │   │   ├── db/                  # D1クエリ（Repository）
│   │   │   └── http/                # DiscordWebhook/外部呼出
│   │   ├── workflows/               # Mastraアダプタ（Edge互換）
│   │   └── types/                   # 共通型（DTO, Env bindings）
│   ├── wrangler.toml                # バインディング（D1/Secrets）
│   └── .dev.vars                    # ローカル開発用（サンプル値のみ）
│
├── mastra/
│   ├── workflows/
│   │   ├── ingest.aimlabs.ts        # parse → validate → enrich → persist
│   │   ├── ingest.kovaaks.ts
│   │   └── post.ingest.summarize.ts # セッション要約/アドバイス起動
│   ├── deployer.json                # CloudflareDeployer 設定
│   └── README.md
│
├── db/
│   ├── migrations/
│   │   └── d1/
│   │       ├── 0001_init.sql        # users/sources/scores/conversations/advice
│   │       └── 0002_indexes.sql
│   └── seeds/
│
├── exporters/
│   └── go/
│       └── aimcoach-exporter/       # Windows向けCLI（Go）
│           ├── cmd/aimcoach-exporter/main.go
│           ├── internal/{config,model,kovaaks,send}
│           ├── go.mod
│           └── README.md
│
├── docs/                            # 追加ドキュメント（API仕様/ER図など）
├── output-sample/                   # 参考用サンプルデータ（本番未使用）
└── .spec-workflow/                  # Steering/Spec ワークフロー資産
```

## Naming Conventions

### Files
- TypeScript: `kebab-case.ts`、テストは `*.test.ts`
- Go: パッケージ小文字、ファイルは `snake_case.go`、テストは `*_test.go`
- SQL: `####_name.sql`（4桁連番）

### Code
- TS 型/インターフェース/Enum: PascalCase（例: `NormalizedRecord`, `IngestResult`）
- 関数/変数: camelCase、定数は UPPER_SNAKE_CASE
- ルート名: `kebab-case`（例: `/ingest/aimlabs`）

## Import Patterns

### TypeScript
1. 外部依存（zod, hono 等）
2. エイリアス（`@/lib/*`, `@/routes/*`）
3. 相対（同ディレクトリ内）
- tsconfig の `paths` で `@/*` を `workers/src/*` にマッピング

### Go (exporter)
1. 標準ライブラリ
2. サードパーティ（`modernc.org/sqlite` 等）
3. 内部パッケージ（`github.com/.../internal/*`）

## Code Structure Patterns

### Module/Class Organization（TS）
1. import
2. 定数/設定
3. 型定義（zod スキーマ含む）
4. 実装（ハンドラ/サービス）
5. エクスポート

### Function Organization
- 入力バリデーション → コア処理 → 永続化/外部呼出 → 例外を `Result` 型に正規化

### File Organization Principles
- 1ファイル=1責務。巨大化時は `lib/*` に抽出

## Code Organization Principles
1. Single Responsibility / Modularity / Testability / Consistency を順守
2. Workers は Edge 互換依存のみ使用（Node 専用 API は禁止）
3. 共有型 `NormalizedRecord` を Workers/Exporter で整合

## Module Boundaries
- API層（routes）→ バリデーション（validation）→ 永続化（db repo）→ ワークフロー（workflows）
- Exporter: 入力（AimLabs/Kovaaks adapters）→ 正規化（model）→ 送信（transport）
- 依存方向は上位→下位のみ。下位から上位への参照は禁止

## Code Size Guidelines
- TS ファイル上限 ~400行、関数 ~60行目安
- Go ファイル上限 ~400行、関数 ~60行目安
- 複雑度が増す場合は分割/抽出

## Dashboard/Monitoring Structure
- Workers: Pino 互換の軽量ロガー、requestId（CF Ray ID）を MDC に保存
- エラーは構造化 JSON。Miniflare 環境でも同一形式を維持

## Documentation Standards
- 各主要ディレクトリに README.md を配置
- API ルートは docs/ にエンドポイント仕様（例・エラーパターン・レート制御）
- マイグレーションは各SQLに目的/ロールバック方針をコメント

## Tooling / Conventions

### Linters / Formatters
- TypeScript（Workers）
  - Linter/Formatter: Biome（lint+format統合）
  - 設定ファイル: `workers/biome.jsonc`
  - 実行: `biome check --write`（CIは`biome ci`）
- Go（exporter）
  - Linter: golangci-lint（`golangci.yml` をリポジトリ直下 or `exporters/go/` 配下に配置）
  - Formatter: `gofmt`/`goimports`

### Git Commit 規約（Conventional Commits）
- フォーマット: `<type>(scope)!: <subject>`
- type: `feat` | `fix` | `docs` | `style` | `refactor` | `perf` | `test` | `build` | `ci` | `chore` | `revert`
- ルール:
  - 英語の命令形で要約（日本語補足は本文/フッターで可）
  - subjectは50文字程度を目安、本文は72桁で折り返し
  - Breaking変更は `!` と `BREAKING CHANGE:` フッターで明示
- 例:
  - `feat(exporter): add kovaaks CSV normalization`
  - `fix(workers): reject invalid metrics with 400`

### Commit 粒度
- 1コミット=1つの論理的変更（機能追加/修正/リファクタ等）
- 原則としてビルド・lint・テストが通る単位で分割
- 大規模変更は事前に下準備（型/リネーム）→ 本体実装 → 仕上げ の複数コミットに分離

### 言語/ツールのバージョン管理（mise）
- ルートに `mise.toml`（または `.tool-versions`）を配置し、開発環境の言語/ツールを固定
- 推奨設定（例）:
  - Node.js: `>=20.9`（Workers開発・ツール用）
  - Go: `1.22.x`（exporter）
  - Deno/Bun: 必要に応じて追記
- 運用: 開発前に `mise install` を実行。CIでも `mise run` / `mise install` を用いて同一バージョンを使用

### Git ブランチ戦略
- モデル: trunk-based（デフォルトブランチ `main`）
- ブランチ命名:
  - 機能: `feat/<short-id>-<kebab-desc>`
  - 修正: `fix/<short-id>-<kebab-desc>`
  - ホットフィックス: `hotfix/<issue>`
- 運用ルール:
  - 変更は短命ブランチで作業し、PRで `main` にマージ
  - マージ方式は基本 `squash and merge`（履歴を読みやすく保つ）
  - CI（lint/test/build）グリーン必須、少なくとも1名レビュー
  - リリースはタグ `vX.Y.Z` を付与（必要に応じて release ブランチを切る）
