import { DiscordHono } from 'discord-hono';
import type { Env, DiscordHonoContext } from './types';
import { CONFIG } from './config';
import { initializeEnvironment, type EnvironmentConfig } from './environment';

/**
 * Shared environment validation and configuration handler
 */
function withEnvironment(
  handler: (envConfig: EnvironmentConfig) => string
) {
  return (c: DiscordHonoContext) => {
    try {
      const envConfig = initializeEnvironment(c.env);
      return c.res(handler(envConfig));
    } catch (error) {
      console.error('Environment validation failed:', error);
      return c.res('⚠️ システム設定エラーが発生しました。管理者に連絡してください。');
    }
  };
}

/**
 * Discord Bot Interface for AimCoach
 * Cloudflare Workers implementation using discord-hono
 *
 * discord-hono automatically handles:
 * - Discord signature verification
 * - Interaction routing
 * - Response formatting
 */
const app = new DiscordHono<Env>()
  .command('help', withEnvironment(() => CONFIG.MESSAGES.HELP))
  .command('analyze', withEnvironment((envConfig) => {
    // Analytics service URL is available in envConfig.services.analyticsEngineUrl
    return CONFIG.MESSAGES.ANALYZE_PLACEHOLDER;
  }))
  .command('training', withEnvironment((envConfig) => {
    // External service configurations are available in envConfig
    return CONFIG.MESSAGES.TRAINING_PLACEHOLDER;
  }))
  .command('status', withEnvironment((envConfig) => {
    // User session and bot config bindings are available in envConfig.bindings
    return CONFIG.MESSAGES.STATUS_PLACEHOLDER;
  }));

export default app;