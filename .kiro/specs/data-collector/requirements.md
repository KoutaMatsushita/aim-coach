# Requirements Document

## Introduction
Data Collectorは、Windowsローカル環境でAimLabとKovaaaksのスコアデータを自動収集し、クラウドへ同期するシステムです。冪等性を保証してデータの重複や不整合を防ぎ、Discordユーザーとゲームスコアをシームレスに紐づけることで、analytics-engineでの分析とdiscord-bot-interfaceでの対話を支援します。

## Requirements

### Requirement 1: ローカルスコアデータ収集
**目的:** FPSプレイヤーとして、自分のAimLabとKovaaaksのスコアデータが自動的に収集され、手動でのデータ入力作業なしに分析システムで活用したい。そうすることで効率的にトレーニング成果を追跡できる。

#### 受入条件
1. WHEN AimLabのスコアファイルが生成された場合 THEN Data Collector SHALL 自動的にファイルを検出してスコアデータを読み込む
2. WHEN Kovaaaksの「[シナリオ名] - Challenge - YYYY.MM.DD-HH.MM.SS Stats.csv」形式ファイルが生成された場合 THEN Data Collector SHALL CSVファイルを解析してスコアデータを抽出する
3. WHEN Kovaaaksファイルから統計データを抽出する場合 THEN Data Collector SHALL 個別キル記録、武器統計、総合スコア、設定情報を正規化して取得する
4. WHEN 新しいスコアデータが検出された場合 THEN Data Collector SHALL ハッシュ値を基にした重複チェックとデータ完全性を検証してから処理を実行する
5. WHERE ファイルシステム監視を実行する場合 THE Data Collector SHALL 低CPU使用率でバックグラウンド動作を継続する

### Requirement 2: Discordユーザー紐づけ機能
**目的:** システム管理者として、Windowsローカルで収集されたスコアデータを適切なDiscordユーザーと関連付けたい。そうすることでユーザーの対話履歴と分析結果を統合的に管理できる。

#### 受入条件
1. WHEN 初回起動時にユーザー認証が要求された場合 THEN Data Collector SHALL Discord OAuth2フローを通じてユーザー認証を実行する
2. WHEN Discord認証が完了した場合 THEN Data Collector SHALL ユーザーIDを暗号化してローカルに安全に保存する
3. IF ユーザーが複数のDiscordアカウントを持つ場合 THEN Data Collector SHALL 主要アカウントの選択機能を提供する
4. WHILE アプリケーションが動作している場合 THE Data Collector SHALL 認証状態を維持してユーザー再認証を最小化する

### Requirement 3: 冪等性を保証するクラウド同期
**目的:** 開発者として、ネットワーク障害や重複処理による不正なデータ状態を防ぎたい。そうすることでデータの信頼性を保証し、システム全体の安定性を確保できる。

#### 受入条件
1. WHEN 同一スコアデータが複数回送信された場合 THEN Data Collector SHALL 重複を検出して冗長な処理を回避する
2. WHEN ネットワーク接続が回復した場合 THEN Data Collector SHALL 未送信データを順次同期して整合性を保つ
3. IF 同期処理中にエラーが発生した場合 THEN Data Collector SHALL 指数バックオフでリトライ処理を実行する
4. WHILE オフライン状態が継続する場合 THE Data Collector SHALL ローカルキューにデータを蓄積して接続回復を待機する

### Requirement 4: データ形式の統一と検証
**目的:** システム統合担当者として、異なるゲームソースからのデータを統一的に処理したい。そうすることでanalytics-engineでの分析精度を向上させることができる。

#### 受入条件
1. WHEN AimLabのデータを処理する場合 THEN Data Collector SHALL SQLiteダンプまたはJSON形式から統一スキーマに正規化してバリデーションを実行する
2. WHEN KovaaaksのCSVデータを処理する場合 THEN Data Collector SHALL 複数セクション構造（キル記録、武器統計、総合統計、設定）を解析して構造化データに変換する
3. WHEN Kovaaaksの詳細データを抽出する場合 THEN Data Collector SHALL 精度、反応時間、武器別統計、シナリオハッシュ、ゲームバージョンを含む包括的メタデータを取得する
4. IF データ形式に不整合が検出された場合 THEN Data Collector SHALL 具体的なエラー箇所とデータソースを記録して管理者に通知する
5. WHERE データ品質チェックを実行する場合 THE Data Collector SHALL 異常値検出、スコア妥当性検証、タイムスタンプ整合性チェックを適用する

### Requirement 5: セキュリティとプライバシー保護
**目的:** ユーザーとして、自分のゲームデータと個人情報が適切に保護され、不正アクセスから守られることを確認したい。そうすることで安心してシステムを利用できる。

#### 受入条件
1. WHEN 機密データを保存する場合 THEN Data Collector SHALL AES-256暗号化を使用してローカルストレージを保護する
2. WHEN クラウドにデータを送信する場合 THEN Data Collector SHALL HTTPS通信と証明書検証を強制する
3. IF 不正なアクセス試行が検出された場合 THEN Data Collector SHALL アクセスを拒否してセキュリティログを記録する
4. WHILE 個人情報を処理する場合 THE Data Collector SHALL 最小権限原則に従ってデータアクセスを制限する

### Requirement 6: ファイルシステム監視とパス管理
**目的:** Windows環境のユーザーとして、AimLabとKovaaaksの標準インストール場所からスコアファイルが自動発見されることを期待する。そうすることで手動設定なしにデータ収集を開始できる。

#### 受入条件
1. WHEN 初回起動時にファイルパス検出を実行する場合 THEN Data Collector SHALL Kovaaaksの標準パス「\SteamApps\common\FPSAimTrainer\FPSAimTrainer\stats\」を自動検出する
2. WHEN AimLabのインストールを検出する場合 THEN Data Collector SHALL Steam、Epic Games、直接インストールの一般的なパスを順次検索する
3. IF 標準パスでファイルが見つからない場合 THEN Data Collector SHALL ユーザーに対してカスタムパス設定のオプションを提供する
4. WHEN バックグラウンド監視モードが有効な場合 THEN Data Collector SHALL Windows File Systemイベントを使用して効率的なリアルタイム検出を実現する
5. IF システム負荷やゲーム影響が懸念される場合 THEN Data Collector SHALL 手動同期モードに切り替えてユーザー操作による都度アップロードを可能にする

### Requirement 7: パフォーマンスとゲーム影響最小化
**目的:** ゲーマーとして、エイムトレーニング中にData Collectorが一切のゲーム体験阻害を引き起こさないことを最優先とする。そうすることで集中してトレーニングに取り組むことができる。

#### 受入条件
1. WHEN ゲームプロセスが検出された場合 THEN Data Collector SHALL 自動的に最小モードに移行してリソース使用を極限まで抑制する
2. WHEN 手動同期モードが選択された場合 THEN Data Collector SHALL ユーザーの明示的な操作でのみデータアップロードを実行する
3. IF バックグラウンド監視でシステム負荷が5%を超過した場合 THEN Data Collector SHALL 自動的に監視を一時停止してゲーム終了後に処理を再開する
4. WHEN ユーザーがゲーム終了後にアプリケーションを開いた場合 THEN Data Collector SHALL 新しいスコアファイルを自動スキャンして同期待ちリストを表示する
5. WHERE 手動アップロード実行時に THE Data Collector SHALL 10秒以内にクラウド同期を完了してユーザーフィードバックを提供する
6. IF analytics-engineからデータ要求を受信した場合 THEN Data Collector SHALL ローカルキャッシュから5秒以内に最新のスコアデータを提供する

### Requirement 8: 運用モード選択とユーザビリティ
**目的:** ユーザーとして、自分のハードウェア環境とゲーム優先度に応じてData Collectorの動作モードを選択したい。そうすることで最適なバランスでデータ収集を利用できる。

#### 受入条件
1. WHEN 初回セットアップ時に運用モード選択が要求された場合 THEN Data Collector SHALL 「リアルタイム監視」「手動同期」「ハイブリッド」の選択肢を提供する
2. WHEN リアルタイム監視モードが選択された場合 THEN Data Collector SHALL 継続的なファイル監視と自動アップロードを実行する
3. WHEN 手動同期モードが選択された場合 THEN Data Collector SHALL ユーザー操作時のみのスキャンとアップロードを実行する
4. WHEN ハイブリッドモードが選択された場合 THEN Data Collector SHALL ゲーム検出時は休止し、ゲーム終了後に自動処理を実行する
5. WHERE 設定変更が要求された場合 THE Data Collector SHALL 運用モードの即座切り替えと設定保存を可能にする

### Requirement 9: エラーハンドリングと可用性
**目的:** システム運用者として、Data Collectorが様々な障害状況に対して適切に対応し、データ損失を防ぎたい。そうすることでサービスの信頼性を確保できる。

#### 受入条件
1. WHEN ファイルアクセスエラーが発生した場合 THEN Data Collector SHALL 権限チェックとパス検証を実行してエラー原因を特定する
2. WHEN クラウド接続が失敗した場合 THEN Data Collector SHALL ローカルバックアップを作成してデータ損失を防ぐ
3. IF アプリケーションクラッシュが発生した場合 THEN Data Collector SHALL 自動復旧機能により処理を再開する
4. WHILE システム監視を実行する場合 THE Data Collector SHALL ヘルスチェック情報を定期的にクラウドに送信する