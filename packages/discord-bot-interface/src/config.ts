/**
 * Configuration constants for the Discord Bot Interface
 */
export const CONFIG = {
  BOT: {
    NAME: 'AimCoach Discord Bot',
    VERSION: '1.0.0',
    RESPONSE_TIMEOUT: 3000, // Discord timeout limit in milliseconds
    PERFORMANCE_TARGET: 1000, // Target response time for simple commands
  },

  DISCORD: {
    INTERACTION_TYPES: {
      PING: 1,
      APPLICATION_COMMAND: 2,
      MESSAGE_COMPONENT: 3,
      APPLICATION_COMMAND_AUTOCOMPLETE: 4,
      MODAL_SUBMIT: 5,
    },

    RESPONSE_TYPES: {
      PONG: 1,
      CHANNEL_MESSAGE_WITH_SOURCE: 4,
      DEFERRED_CHANNEL_MESSAGE_WITH_SOURCE: 5,
      DEFERRED_UPDATE_MESSAGE: 6,
      UPDATE_MESSAGE: 7,
    },
  },

  COMMANDS: {
    HELP: 'help',
    ANALYZE: 'analyze',
    TRAINING: 'training',
    STATUS: 'status',
  },

  MESSAGES: {
    HELP: `
🎯 **AimCoach Discord Bot - 利用可能なコマンド**

\`/aim-coach help\` - このヘルプを表示
\`/aim-coach analyze\` - スコア分析を開始
\`/aim-coach training\` - パーソナライズドトレーニング推奨
\`/aim-coach status\` - あなたの進捗状況を表示

メンション機能:
@AimCoach に質問を投稿すると、AIが回答します。

詳細: https://github.com/aim-coach/docs
    `,
    ANALYZE_PLACEHOLDER: '🔍 スコア分析機能を開始します...\n（この機能は現在実装中です）',
    TRAINING_PLACEHOLDER: '🏋️ パーソナライズドトレーニング推奨を生成中...\n（この機能は現在実装中です）',
    STATUS_PLACEHOLDER: '📊 あなたの進捗状況を取得中...\n（この機能は現在実装中です）',
    UNKNOWN_COMMAND: '❌ 不明なコマンドです。`/aim-coach help` で利用可能なコマンドを確認してください。',
  },
} as const;