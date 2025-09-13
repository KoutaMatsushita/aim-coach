# Technical Design Document

## Architecture Overview

Data Collectorは、Windowsローカル環境でAimLabとKovaaaksのスコアデータを収集し、Cloudflareインフラストラクチャと同期するデスクトップアプリケーションです。ゲームパフォーマンスへの影響を最小限に抑えることを最優先とし、柔軟な動作モードを提供します。

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Data Collector (Windows)                 │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────── │
│  │   Game Process  │  │  File System    │  │  Discord OAuth │
│  │   Detector      │  │  Monitor        │  │  Integration   │
│  └─────────────────┘  └─────────────────┘  └─────────────── │
│           │                     │                     │     │
│           v                     v                     v     │
│  ┌─────────────────────────────────────────────────────────│ │
│  │            Operating Mode Controller                    │ │
│  └─────────────────────────────────────────────────────────│ │
│           │                     │                     │     │
│           v                     v                     v     │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────── │
│  │   Data Parser   │  │  Local Storage  │  │  Sync Queue   │
│  │   Engine        │  │  Manager        │  │  Manager      │
│  └─────────────────┘  └─────────────────┘  └─────────────── │
└─────────────────────────────────────────────────────────────┘
                               │
                               v (HTTPS/TLS)
┌─────────────────────────────────────────────────────────────┐
│              Cloudflare Infrastructure                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────── │
│  │   Discord Bot   │  │  Analytics      │  │  D1 Database  │
│  │   Interface     │  │  Engine         │  │  Storage      │
│  └─────────────────┘  └─────────────────┘  └─────────────── │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Operating Mode Controller

ゲームの実行状態とユーザー設定に基づいて動作モードを制御する中央コンポーネント。

```typescript
interface OperatingModes {
  REALTIME: 'realtime_monitoring';     // 継続的監視と自動同期
  MANUAL: 'manual_sync';               // ユーザー操作時のみ処理
  HYBRID: 'hybrid_mode';               // ゲーム終了後自動処理
}

interface GameProcessDetector {
  detectGameProcesses(): GameProcess[];
  isGameRunning(): boolean;
  onGameStart(callback: () => void): void;
  onGameEnd(callback: () => void): void;
}
```

**動作ロジック:**
- リアルタイムモード: 継続的なファイル監視 + 即座の同期
- 手動モード: ユーザー操作時のスキャン + アップロード
- ハイブリッドモード: ゲーム検出時は休止、終了後に自動処理

### 2. File System Monitor

Windows File System APIを使用した効率的なファイル監視システム。

```typescript
interface FileSystemMonitor {
  // Kovaaks監視パス
  kovaaaksPath: string; // "\SteamApps\common\FPSAimTrainer\FPSAimTrainer\stats\"

  // AimLab監視パス (複数可能性)
  aimlabPaths: string[]; // Steam, Epic Games, 直接インストール

  // ファイル検出パターン
  kovaaaksPattern: RegExp; // /.*- Challenge - \d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2} Stats\.csv$/
  aimlabPattern: RegExp;   // SQLiteファイルまたはCSVエクスポート

  startMonitoring(): void;
  stopMonitoring(): void;
  scanForNewFiles(): Promise<FileInfo[]>;
}
```

**パフォーマンス最適化:**
- Windows ReadDirectoryChangesW API使用
- フィルタリングによる不要なイベント除外
- バッファリングによるCPU使用率制御
- ゲーム実行時の自動停止機能

### 3. Data Parser Engine

ゲーム固有のデータ形式を統一スキーマに正規化するパーサー。

```typescript
interface DataParser {
  parseKovaaaksCSV(filePath: string): Promise<KovaaaksScore>;
  parseAimLabData(source: string, format: 'sqlite' | 'csv'): Promise<AimLabScore>;
  validateDataIntegrity(data: ScoreData): ValidationResult;
  normalizeToUnifiedSchema(data: RawScoreData): UnifiedScoreData;
}

interface UnifiedScoreData {
  // 共通メタデータ
  playerId: string;           // Discord User ID (暗号化)
  gameSource: 'aimlab' | 'kovaaks';
  timestamp: string;          // ISO 8601
  scenario: string;           // シナリオ名
  gameVersion: string;        // ゲームバージョン

  // 統計データ
  accuracy: number;           // 精度 (0-1)
  averageReactionTime: number; // 平均反応時間 (ms)
  killCount: number;          // キル数
  score: number;              // 総合スコア

  // 詳細データ
  weaponStats: WeaponStatistics[];
  killDetails: KillRecord[];
  gameSettings: GameConfiguration;

  // 検証用
  dataHash: string;           // ハッシュ値 (重複防止)
  fileSize: number;           // ファイルサイズ
  checksum: string;           // チェックサム
}
```

**Kovaaaksデータ構造:**
```typescript
interface KovaaaksScore {
  // メタデータ (ファイル名から抽出)
  scenario: string;           // "devTS Goated NR Static Small"
  timestamp: string;          // "2023.02.26-16.12.41"

  // キル記録セクション
  killRecords: KillRecord[];

  // 武器統計セクション
  weaponStats: {
    weapon: string;           // "MGV2"
    shots: number;            // 637
    hits: number;             // 474
    accuracy: number;         // 0.744
    damageEfficiency: number; // 0.372
  };

  // 総合統計セクション
  overallStats: {
    kills: number;            // 79
    averageTTK: number;       // 0.758224
    score: number;            // 474.0
    fightTime: number;        // 25.466
    scenarioHash: string;     // "d18c9493ea46b307efb25daa310c5f8b"
  };

  // 設定セクション
  gameConfig: {
    sensitivity: number;      // 0.28
    dpi: number;             // 800
    fov: number;             // 103.0
    resolution: string;      // "2560x1440"
    crosshair: string;       // "plus.png"
  };
}
```

### 4. Discord OAuth Integration

セキュアなユーザー認証とDiscord連携機能。

```typescript
interface DiscordAuthManager {
  initiateOAuth(): Promise<AuthorizationURL>;
  handleCallback(code: string): Promise<UserProfile>;
  refreshToken(): Promise<AccessToken>;
  encryptAndStoreCredentials(profile: UserProfile): void;
  getCurrentUser(): Promise<UserProfile | null>;
}

interface UserProfile {
  discordId: string;          // Discord User ID
  username: string;           // 表示用ユーザー名
  accessToken: string;        // OAuth アクセストークン (暗号化保存)
  refreshToken: string;       // リフレッシュトークン (暗号化保存)
  linkedAt: string;           // 連携日時
}
```

**セキュリティ実装:**
- AES-256-GCM暗号化によるローカル認証情報保護
- PKCE (Proof Key for Code Exchange) フロー使用
- トークン自動更新機能
- 複数アカウント対応

### 5. Sync Queue Manager

冪等性を保証するクラウド同期システム。

```typescript
interface SyncQueueManager {
  addToQueue(data: UnifiedScoreData): void;
  processQueue(): Promise<SyncResult[]>;
  retryFailedItems(): Promise<void>;
  getDuplicateHash(data: UnifiedScoreData): string;
  validateServerResponse(response: APIResponse): boolean;
}

interface SyncItem {
  id: string;                 // ユニークID
  data: UnifiedScoreData;     // 同期対象データ
  attempts: number;           // リトライ回数
  lastAttempt: string;        // 最終試行時刻
  status: 'pending' | 'syncing' | 'completed' | 'failed';
  hash: string;               // 重複検出用ハッシュ
}
```

**冪等性保証メカニズム:**
- SHA-256ハッシュによる重複検出
- サーバー側での二重処理防止
- 指数バックオフリトライ戦略
- ローカルキューによるオフライン対応

### 6. Local Storage Manager

セキュアなローカルデータ管理システム。

```typescript
interface LocalStorageManager {
  // 設定管理
  saveUserSettings(settings: UserSettings): void;
  loadUserSettings(): UserSettings;

  // キューデータ管理
  persistSyncQueue(queue: SyncItem[]): void;
  loadSyncQueue(): SyncItem[];

  // 認証データ管理 (暗号化)
  storeEncryptedCredentials(credentials: EncryptedCredentials): void;
  loadEncryptedCredentials(): EncryptedCredentials | null;

  // キャッシュ管理
  cacheProcessedFiles(fileHashes: string[]): void;
  isFileProcessed(hash: string): boolean;
}
```

## Data Flow

### 1. リアルタイム監視フロー

```
ファイル作成イベント → ゲーム実行チェック → パース処理 → 重複検証 → 同期キュー追加 → クラウド送信
```

### 2. 手動同期フロー

```
ユーザー操作 → ディレクトリスキャン → 新規ファイル検出 → バッチ処理 → 進捗表示 → 完了通知
```

### 3. ハイブリッドフロー

```
ゲーム終了検出 → 自動スキャン開始 → バックグラウンド処理 → ユーザー通知
```

## API Integration

### Analytics Engine API

```typescript
interface AnalyticsEngineAPI {
  endpoint: string; // "https://analytics.aim-coach.com/api/v1"

  submitScoreData(data: UnifiedScoreData): Promise<SubmissionResult>;
  validateDataFormat(data: UnifiedScoreData): Promise<ValidationResult>;
  getPlayerProgress(playerId: string): Promise<ProgressData>;
}

interface SubmissionResult {
  success: boolean;
  scoreId: string;            // サーバー側で生成されたID
  duplicateDetected: boolean; // 重複検出フラグ
  validationErrors: string[]; // バリデーションエラー
  processingTime: number;     // 処理時間 (ms)
}
```

### Discord Bot Interface API

```typescript
interface DiscordBotAPI {
  endpoint: string; // "https://discord-bot.aim-coach.com/api/v1"

  notifyScoreUpdate(playerId: string, scoreId: string): Promise<void>;
  getUserPreferences(playerId: string): Promise<UserPreferences>;
  sendDirectMessage(playerId: string, message: string): Promise<void>;
}
```

## Performance Optimization

### 1. ゲーム影響最小化

```typescript
interface GameImpactMinimizer {
  // プロセス監視
  monitorGameProcesses(): void;

  // リソース制限
  limitCPUUsage(maxPercentage: number): void;
  limitMemoryUsage(maxMB: number): void;

  // 優先度制御
  setProcessPriority(priority: 'low' | 'normal'): void;

  // 自動休止
  suspendDuringGameplay(): void;
  resumeAfterGameplay(): void;
}
```

### 2. パフォーマンス指標

- CPU使用率: < 5% (バックグラウンド動作時)
- メモリ使用量: < 50MB (待機時)
- ディスクI/O: < 1MB/min (監視時)
- ネットワーク: バースト送信、継続的通信なし

## Security Implementation

### 1. データ暗号化

```typescript
interface EncryptionManager {
  encryptSensitiveData(data: string): Promise<EncryptedData>;
  decryptSensitiveData(encrypted: EncryptedData): Promise<string>;
  generateDataHash(data: UnifiedScoreData): string;
  verifyDataIntegrity(data: UnifiedScoreData, hash: string): boolean;
}

interface EncryptedData {
  ciphertext: string;         // AES-256-GCM暗号化データ
  iv: string;                 // 初期化ベクトル
  tag: string;                // 認証タグ
  algorithm: 'AES-256-GCM';   // 暗号化アルゴリズム
}
```

### 2. 通信セキュリティ

- HTTPS/TLS 1.3強制
- 証明書ピニング
- リクエスト署名検証
- レート制限対応

## Error Handling

### 1. エラー分類

```typescript
enum ErrorType {
  FILE_ACCESS_ERROR = 'file_access_error',
  NETWORK_ERROR = 'network_error',
  PARSING_ERROR = 'parsing_error',
  AUTHENTICATION_ERROR = 'auth_error',
  VALIDATION_ERROR = 'validation_error'
}

interface ErrorHandler {
  handleError(error: Error, type: ErrorType): void;
  retryWithBackoff(operation: () => Promise<any>, maxRetries: number): Promise<any>;
  logError(error: Error, context: ErrorContext): void;
  notifyUser(error: UserFacingError): void;
}
```

### 2. 復旧戦略

- ファイルアクセス失敗: 権限チェック + 再試行
- ネットワーク失敗: オフラインキュー + 指数バックオフ
- パース失敗: データ検証 + ユーザー通知
- 認証失敗: 自動再認証 + フォールバック

## Technology Stack

### Core Technologies
- **Language**: TypeScript/Node.js
- **UI Framework**: Electron (クロスプラットフォーム対応)
- **State Management**: Zustand
- **File System**: Node.js fs/promises + chokidar
- **Encryption**: Node.js crypto module
- **HTTP Client**: Axios with retry interceptors

### Development Tools
- **Build**: Electron Builder
- **Testing**: Jest + Electron Testing
- **Linting**: ESLint + TypeScript
- **Packaging**: Auto-updater対応

### Deployment
- **Distribution**: GitHub Releases
- **Auto-Update**: electron-updater
- **Installer**: NSIS (Windows)
- **Code Signing**: Windows Authenticode

## Deployment Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    User's Windows PC                        │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────│ │
│  │            Data Collector App                           │ │
│  │  (Electron + TypeScript)                               │ │
│  └─────────────────────────────────────────────────────────│ │
│                               │                             │
│  ┌─────────────────────────────────────────────────────────│ │
│  │              Local Data Storage                         │ │
│  │  - 設定ファイル (暗号化)                                │ │
│  │  - 同期キュー (永続化)                                  │ │
│  │  - 処理済みファイルキャッシュ                            │ │
│  └─────────────────────────────────────────────────────────│ │
└─────────────────────────────────────────────────────────────┘
                               │ HTTPS
                               v
┌─────────────────────────────────────────────────────────────┐
│                Cloudflare Edge Network                      │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────── │
│  │   Workers       │  │   D1 Database   │  │   KV Store    │
│  │   (API Gateway) │  │   (Score Data)  │  │   (Cache)     │
│  └─────────────────┘  └─────────────────┘  └─────────────── │
└─────────────────────────────────────────────────────────────┘
```

この設計により、ゲームパフォーマンスへの影響を最小限に抑えながら、信頼性の高いスコアデータ収集と同期を実現します。