import { DiscordHono } from 'discord-hono';
import type { Env } from './types';
import { CONFIG } from './config';

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
  .command('help', (c) => c.res(CONFIG.MESSAGES.HELP))
  .command('analyze', (c) => c.res(CONFIG.MESSAGES.ANALYZE_PLACEHOLDER))
  .command('training', (c) => c.res(CONFIG.MESSAGES.TRAINING_PLACEHOLDER))
  .command('status', (c) => c.res(CONFIG.MESSAGES.STATUS_PLACEHOLDER));

export default app;