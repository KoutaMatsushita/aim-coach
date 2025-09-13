/**
 * Environment variables interface for Cloudflare Workers
 */
export interface Env {
  // Discord Bot settings
  DISCORD_TOKEN: string;
  DISCORD_PUBLIC_KEY: string;
  DISCORD_APPLICATION_ID: string;

  // External services
  ANALYTICS_ENGINE_URL: string;
  CONVERSATION_ENGINE_URL: string;
  INTERNAL_API_TOKEN: string;

  // Durable Objects
  USER_SESSION: DurableObjectNamespace;

  // KV Storage
  BOT_CONFIG: KVNamespace;
}

/**
 * Standard API response format
 */
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: {
    code: string;
    message: string;
  };
  timestamp: string;
}

/**
 * User session data structure
 */
export interface SessionData {
  userId: string;
  preferences: UserPreferences;
  conversationHistory: ConversationEntry[];
  lastInteraction: Date;
  trainingContext: TrainingContext;
}

export interface UserPreferences {
  language: 'ja' | 'en';
  notificationsEnabled: boolean;
  proactiveMessaging: boolean;
}

export interface ConversationEntry {
  timestamp: Date;
  type: 'user' | 'bot';
  content: string;
  context?: Record<string, unknown>;
}

export interface TrainingContext {
  lastScoreAnalysis?: Date;
  currentGoals: string[];
  weaknesses: string[];
  strengths: string[];
}

/**
 * Discord interaction types
 */
export interface DiscordInteraction {
  id: string;
  type: number;
  data?: SlashCommandData;
  user?: DiscordUser;
  member?: DiscordMember;
  token: string;
}

export interface SlashCommandData {
  id: string;
  name: string;
  options?: CommandOption[];
}

export interface CommandOption {
  name: string;
  type: number;
  value?: string | number | boolean;
}

export interface DiscordUser {
  id: string;
  username: string;
  discriminator: string;
  avatar?: string;
}

export interface DiscordMember {
  user?: DiscordUser;
  nick?: string;
  roles: string[];
}

/**
 * Error handling
 */
export class AimCoachError extends Error {
  constructor(
    public code: string,
    public message: string,
    public context?: Record<string, unknown>
  ) {
    super(message);
    this.name = 'AimCoachError';
  }
}

export enum ErrorCode {
  DISCORD_AUTH_FAILED = 'DISCORD_AUTH_FAILED',
  EXTERNAL_SERVICE_TIMEOUT = 'EXTERNAL_SERVICE_TIMEOUT',
  SESSION_CORRUPTED = 'SESSION_CORRUPTED',
  RATE_LIMIT_EXCEEDED = 'RATE_LIMIT_EXCEEDED',
  INVALID_COMMAND_PARAMETERS = 'INVALID_COMMAND_PARAMETERS',
  USER_NOT_FOUND = 'USER_NOT_FOUND',
}