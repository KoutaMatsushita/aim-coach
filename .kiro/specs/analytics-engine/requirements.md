# Requirements Document

## Introduction
Analytics Engineは、Cloudflare D1とWorkersベースのスコア分析エンジンとして、Gemini APIを活用してFPSプレイヤーのエイムスコアを分析し、パーソナライズされたアドバイスとトレーニング提案を提供するシステムです。過去の対話履歴を考慮することで、各ユーザーに最適化された指導を実現します。

## Requirements

### Requirement 1: スコア分析とアドバイス生成
**目的:** FPSプレイヤーとして、自分のエイムスコアが専門的に分析され、個別最適化されたアドバイスを受けたい。そうすることで効率的にエイム技術を向上させることができる。

#### 受入条件
1. WHEN AimLabまたはKovaaaksのスコアデータが提供された場合 THEN Analytics Engine SHALL Gemini APIを使用してスコアを分析する
2. WHEN スコア分析が完了した場合 THEN Analytics Engine SHALL ユーザーの強み・弱み・改善点を特定したレポートを生成する
3. WHEN 分析結果が生成された場合 THEN Analytics Engine SHALL 日本語で理解しやすいアドバイスを提供する
4. WHERE ユーザーの技術レベルや過去の成績を考慮する場合 THE Analytics Engine SHALL 個人の成長段階に適したアドバイスを生成する

### Requirement 2: パーソナライゼーション機能
**目的:** ユーザーとして、自分の過去の対話や成績履歴を基にした個別指導を受けたい。そうすることで一般的でないより具体的で効果的なトレーニング提案を得ることができる。

#### 受入条件
1. WHEN ユーザーの過去の対話履歴が利用可能な場合 THEN Analytics Engine SHALL discord-bot-interfaceから履歴データを取得する
2. WHEN 複数回の分析履歴が存在する場合 THEN Analytics Engine SHALL 成績の推移と改善傾向を分析する
3. IF ユーザーが特定の弱点を繰り返し指摘されている場合 THEN Analytics Engine SHALL より詳細な改善プランを提案する
4. WHILE 新しいスコアデータが追加される場合 THE Analytics Engine SHALL 継続的に学習モデルを更新する

### Requirement 3: トレーニング推奨システム
**目的:** プレイヤーとして、自分の弱点を克服するための具体的なトレーニングメニューと練習方法を知りたい。そうすることで目標を持って効率的に練習することができる。

#### 受入条件
1. WHEN スコア分析で特定の弱点が検出された場合 THEN Analytics Engine SHALL 対応するトレーニングメニューを推奨する
2. WHEN トレーニング推奨を生成する場合 THEN Analytics Engine SHALL AimLabやKovaaaksの具体的なシナリオ名を含める
3. IF ユーザーの技術レベルが初心者の場合 THEN Analytics Engine SHALL 基礎的なトレーニングから段階的に提案する
4. WHERE 上級者向けの推奨を行う場合 THE Analytics Engine SHALL より高度で専門的なトレーニング内容を提供する

### Requirement 4: データ永続化と履歴管理
**目的:** システム管理者として、ユーザーのスコアデータと分析結果を安全に保存し、継続的な改善提案のために活用したい。そうすることでサービス品質を向上させることができる。

#### 受入条件
1. WHEN スコアデータが受信された場合 THEN Analytics Engine SHALL Cloudflare D1データベースに安全に保存する
2. WHEN 分析結果が生成された場合 THEN Analytics Engine SHALL 結果をユーザーIDと関連付けて保存する
3. WHILE データの保存処理を実行する場合 THE Analytics Engine SHALL データの整合性と暗号化を保証する
4. IF 過去の分析データを参照する場合 THEN Analytics Engine SHALL 効率的なクエリでデータを取得する

### Requirement 5: API統合とパフォーマンス
**目的:** 開発者として、Analytics Engineがdiscord-bot-interfaceやconversation-engineと円滑に連携し、ユーザーに快適な体験を提供したい。そうすることでシステム全体の価値を最大化することができる。

#### 受入条件
1. WHEN discord-bot-interfaceからAPIリクエストを受信した場合 THEN Analytics Engine SHALL 30秒以内に分析結果を返す
2. WHEN conversation-engineから会話中のスコア分析要求を受信した場合 THEN Analytics Engine SHALL リアルタイムで分析結果を提供する
3. WHEN Gemini APIとの通信エラーが発生した場合 THEN Analytics Engine SHALL 適切なエラーハンドリングとリトライ機構を実行する
4. WHILE 複数のリクエストが同時に処理される場合 THE Analytics Engine SHALL Cloudflare Workersの制限内で効率的に処理する
5. WHERE システム負荷が高い場合 THE Analytics Engine SHALL レスポンス品質を維持しながら処理を継続する
6. IF conversation-engineから会話の文脈でスコア分析が必要な場合 THEN Analytics Engine SHALL 会話に適した形式で分析結果を返す

### Requirement 6: セキュリティとプライバシー
**目的:** ユーザーとして、自分のゲームデータと対話履歴が適切に保護され、安全に処理されることを確認したい。そうすることで安心してサービスを利用することができる。

#### 受入条件
1. WHEN ユーザーデータを処理する場合 THEN Analytics Engine SHALL 個人情報保護の原則に従って処理する
2. WHEN データをGemini APIに送信する場合 THEN Analytics Engine SHALL 必要最小限の情報のみを送信する
3. IF データベースにアクセスする場合 THEN Analytics Engine SHALL 適切な認証と承認を実行する
4. WHILE ログを記録する場合 THE Analytics Engine SHALL 機密情報の漏洩を防ぐ