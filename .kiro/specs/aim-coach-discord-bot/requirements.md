# Requirements Document

## はじめに
AimCoachは、AimLabとKovaaaksのスコアデータを収集・分析し、LLMベースの対話型インターフェースを通じて個別ユーザーに最適化されたエイムトレーニング指導を提供するクラウドネイティブシステムです。ローカルデータ収集、クラウドデータ処理・保存、Cloudflare Workers上のDiscord Bot、対話履歴管理を統合し、スケーラブルで効率的なプレイヤー技術向上支援を実現します。

## Requirements

### 要件 1: ローカルスコア収集とクラウド同期（Data Collector + Upload Service）
**目標:** エイムトレーニングプレイヤーとして、AimLabとKovaaaksの成績をローカルで自動収集し、クラウドに安全にアップロードできるようにしたい。そうすることで手動でのデータ入力作業を削減し、どこからでもアクセス可能な分析基盤を構築できる。

#### 受入基準
1. WHEN ユーザーがローカルファイル監視を設定する THEN Data Collectorは AimLab TaskData.csv と Kovaaks Stats.csv の変更を自動検出する
2. WHEN 新しいスコアデータが検出された場合 THEN Data Collectorは標準化されたJSONフォーマットでローカル保存しクラウドにアップロードする
3. IF ローカルファイルアクセスに問題がある場合 THEN Data Collectorは詳細なエラーログを出力しクラウド同期を一時停止する
4. WHEN ユーザーが手動収集を実行する場合 THEN Data Collectorは指定された期間のデータを収集しバッチアップロードする
5. WHILE 収集・アップロード処理が実行中の場合 THE Data Collectorは進行状況をローカルUIで表示する
6. WHERE ネットワーク接続が不安定な場合 THE Data Collectorは自動リトライとオフライン対応を実行する

### 要件 2: クラウドベース分析エンジン（Analytics Engine on Cloudflare）
**目標:** エイムトレーニングプレイヤーとして、クラウドに保存されたスコアデータに基づいた高速で詳細な分析と改善提案を受けたい。そうすることで弱点を特定し、効率的に技術向上できる。

#### 受入基準
1. WHEN 分析エンジンがクラウドDBの新データを検出する THEN Cloudflare Workers上で統計的分析を実行する
2. IF ユーザーのスコアに改善傾向が見られる場合 THEN Analytics Engineは進歩レポートと継続戦略をCloudflare D1に保存する
3. IF ユーザーのスコアに停滞や悪化が見られる場合 THEN Analytics Engineは改善アドバイスレポートを生成しキャッシュする
4. WHEN 複数のゲームタイプデータが利用可能な場合 THEN Analytics Engineはサーバーレス環境で総合パフォーマンス評価を実行する
5. WHERE 特定の技術領域で弱点が特定された場合 THE Analytics Engineは該当領域に特化したトレーニング推奨を生成する
6. WHILE 高負荷時の場合 THE Cloudflare Workers は自動スケーリングで複数リクエストを並行処理する

### 要件 3: LLM対話エンジン（Conversation Engine）
**目標:** エイムトレーニングプレイヤーとして、自然言語での対話を通じてスコアデータと個人的な状況を統合した、パーソナライズされたアドバイスを受けたい。そうすることでより具体的で実行可能なトレーニング改善提案を得られる。

#### 受入基準
1. WHEN ユーザーが自然言語でエイム関連の質問をする THEN Conversation Engineはスコアデータを参照した回答を生成する
2. WHEN ユーザーが今日の調子や状況について言及する THEN Conversation Engineは会話履歴として記録し今後の提案に反映する
3. IF ユーザーが特定の技術的な悩みを相談する場合 THEN Conversation Engineはスコア傾向と会話コンテキストを組み合わせた具体的アドバイスを提供する
4. WHEN ユーザーが前回の提案について感想を述べる場合 THEN Conversation Engineは提案の有効性を学習し改善に活用する
5. WHILE 対話が継続中の場合 THE Conversation Engineは会話コンテキストとユーザーの感情状態を維持する
6. WHERE 技術的な限界で回答できない場合 THE Conversation Engineは制限を明確に伝え代替手段を提示する

### 要件 4: Cloudflare Workers Discord Bot
**目標:** Discordユーザーとして、高速で安定したサーバーレス環境で動作するbotを通じて、直感的なコマンドと自然な会話で分析結果やトレーニング推奨にアクセスしたい。そうすることで日常的なコミュニケーション環境でコーチング機能を利用できる。

#### 受入基準
1. WHEN ユーザーがslash command "/help"を実行する THEN Cloudflare Workers上のDiscord Botは利用可能な全コマンドの説明を表示する
2. WHEN ユーザーが"/analysis"コマンドを実行する THEN Discord BotはCloudflare D1から最新分析結果を取得し表示する
3. WHEN ユーザーが"/training"コマンドを実行する THEN Discord BotはConversation Engineによるパーソナライズされたトレーニング推奨を表示する
4. WHEN ユーザーがbotにメンションで自然言語メッセージを送信する THEN Discord BotはWebhook経由でConversation Engineに転送し会話として処理する
5. WHEN ユーザーが"/sync"コマンドを実行する THEN Discord Botはローカル収集システムの同期状況をCloudflare KVから表示する
6. IF Conversation EngineまたはLLM APIが利用不可能な場合 THEN Discord Botは基本分析結果による代替応答を表示する
7. WHERE 複数ユーザーが同時にbotを使用する場合 THE Cloudflare Workersは各ユーザーのセッションをDurable Objectsで独立管理する

### 要件 5: ツール間通信とデータ連携
**目標:** システム管理者として、各独立ツール間（Data Collector, Analytics Engine, Conversation Engine, Discord Bot）でデータが適切に連携され、統合されたワークフローが実現できるようにしたい。そうすることでモジュラーアーキテクチャの利点を活かしながら統一されたユーザー体験を提供できる。

#### 受入基準
1. WHEN Data Collectorが新しいスコアを収集する THEN 標準化されたデータフォーマットでAnalytics EngineとConversation Engineに通知する
2. WHEN Analytics Engineが分析を完了する THEN 結果をConversation Engineとインターフェースがアクセス可能な形式で保存する
3. WHEN Conversation Engineが会話を処理する THEN スコアデータと分析結果を参照し会話履歴を更新する
4. IF いずれかのツールが利用不可能な場合 THEN 他のツールは適切なエラーハンドリングを実行する
5. WHEN ユーザーがDiscord Botを通じてリクエストする場合 THEN Discord Botは適切なツールにリクエストを転送する
6. WHILE データ同期が実行中の場合 THE システムは各ツール間の整合性を維持する
7. WHERE 設定変更が発生した場合 THE 全ツールに変更が適切に伝播される

### 要件 6: 代替インターフェース（Optional）
**目標:** ユーザーとして、Discordインターフェース以外からも分析結果やトレーニング推奨にアクセスできるようにしたい。そうすることで柔軟な利用方法を選択できる。

#### 受入基準
1. WHEN ユーザーがコマンドライン実行を選択する THEN CLIツールが分析結果を表示する
2. WHEN ユーザーがWebインターフェースを選択する THEN ブラウザで詳細な分析レポートとチャット履歴を閲覧できる
3. IF ユーザーがファイル出力を選択する場合 THEN 分析結果と会話履歴をJSON/CSVフォーマットで出力する
4. WHERE 複数のインターフェースが同時利用される場合 THE 各インターフェースは同一のデータソースを参照する

### 要件 7: 基本的なデータ保護（個人利用向け）
**目標:** 個人利用者として、最低限のデータ保護が実装されていることを確認したい。そうすることで基本的な安全性を保ちながらシンプルなシステム構成を維持できる。

#### 受入基準
1. WHEN ユーザーがファイルパスを設定する場合 THEN システムは読み取り専用アクセスのみを使用する
2. IF ユーザーがデータ削除を要求した場合 THEN システムは収集データ、会話履歴、ユーザー設定を削除する
3. WHILE データを保存している間 THE システムはローカル暗号化ストレージを使用する
4. WHERE 機密情報（パス情報、会話履歴等）を扱う場合 THE システムは設定ファイルのアクセス権限を制限する
5. WHEN LLM APIキーを管理する場合 THEN システムは環境変数または暗号化設定ファイルで安全に保存する

### 要件 8: 運用とメンテナンス（簡素化）
**目標:** 個人利用者として、複雑な運用手順なしでシステムを維持できるようにしたい。そうすることで技術的負担を最小化しながら安定した動作を確保できる。

#### 受入基準
1. WHEN 致命的なエラーが発生した場合 THEN システムはログファイルに詳細を記録する
2. IF ツール間通信が失敗した場合 THEN 各ツールは独立して動作を継続する
3. IF LLM APIが利用不可能または制限に達した場合 THEN システムは基本分析結果での代替応答を提供する
4. WHILE システムが動作中の場合 THE 各ツールは基本的なヘルスチェック機能を提供する
5. WHERE 設定の問題が発生した場合 THE システムはわかりやすいエラーメッセージを表示する
6. WHEN アップデートが利用可能な場合 THEN システムは通知機能を提供する（オプション）