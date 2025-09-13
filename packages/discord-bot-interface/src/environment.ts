import type { Env } from './types';

/**
 * Environment validation result
 */
export interface ValidationResult {
  valid: boolean;
  errors: string[];
}

/**
 * Structured environment configuration
 */
export interface EnvironmentConfig {
  discord: {
    token: string;
    publicKey: string;
    applicationId: string;
  };
  services: {
    analyticsEngineUrl: string;
    conversationEngineUrl: string;
    internalApiToken: string;
  };
  bindings: {
    userSession: DurableObjectNamespace;
    botConfig: KVNamespace;
  };
}

/**
 * Environment types for different deployment contexts
 */
export type EnvironmentType = 'development' | 'staging' | 'production';

/**
 * Validate environment configuration
 */
export function validateEnvironment(
  env: Partial<Env>,
  environmentType: EnvironmentType = 'development'
): ValidationResult {
  const errors: string[] = [];

  // Discord Bot configuration validation
  if (!env.DISCORD_TOKEN) {
    errors.push('DISCORD_TOKEN is required');
  }

  if (!env.DISCORD_PUBLIC_KEY) {
    errors.push('DISCORD_PUBLIC_KEY is required');
  }

  if (!env.DISCORD_APPLICATION_ID) {
    errors.push('DISCORD_APPLICATION_ID is required');
  }

  // External services validation
  if (!env.ANALYTICS_ENGINE_URL) {
    errors.push('ANALYTICS_ENGINE_URL is required');
  } else {
    // Validate URL format
    try {
      const url = new URL(env.ANALYTICS_ENGINE_URL);
      // In production, require HTTPS
      if (environmentType === 'production' && url.protocol !== 'https:') {
        errors.push('ANALYTICS_ENGINE_URL must use HTTPS in production');
      }
    } catch {
      errors.push('ANALYTICS_ENGINE_URL must be a valid URL');
    }
  }

  if (!env.CONVERSATION_ENGINE_URL) {
    errors.push('CONVERSATION_ENGINE_URL is required');
  } else {
    // Validate URL format
    try {
      const url = new URL(env.CONVERSATION_ENGINE_URL);
      // In production, require HTTPS
      if (environmentType === 'production' && url.protocol !== 'https:') {
        errors.push('CONVERSATION_ENGINE_URL must use HTTPS in production');
      }
    } catch {
      errors.push('CONVERSATION_ENGINE_URL must be a valid URL');
    }
  }

  if (!env.INTERNAL_API_TOKEN) {
    errors.push('INTERNAL_API_TOKEN is required');
  }

  // Cloudflare bindings validation
  if (!env.USER_SESSION) {
    errors.push('USER_SESSION Durable Object binding is required');
  }

  if (!env.BOT_CONFIG) {
    errors.push('BOT_CONFIG KV namespace binding is required');
  }

  return {
    valid: errors.length === 0,
    errors,
  };
}

/**
 * Get structured environment configuration
 */
export function getEnvironmentConfig(env: Env): EnvironmentConfig {
  return {
    discord: {
      token: env.DISCORD_TOKEN,
      publicKey: env.DISCORD_PUBLIC_KEY,
      applicationId: env.DISCORD_APPLICATION_ID,
    },
    services: {
      analyticsEngineUrl: env.ANALYTICS_ENGINE_URL,
      conversationEngineUrl: env.CONVERSATION_ENGINE_URL,
      internalApiToken: env.INTERNAL_API_TOKEN,
    },
    bindings: {
      userSession: env.USER_SESSION,
      botConfig: env.BOT_CONFIG,
    },
  };
}

/**
 * Detect environment type from various indicators
 */
export function detectEnvironment(env: Env): EnvironmentType {
  // Check for environment-specific indicators
  if (env.ANALYTICS_ENGINE_URL?.includes('localhost') ||
      env.CONVERSATION_ENGINE_URL?.includes('localhost')) {
    return 'development';
  }

  if (env.ANALYTICS_ENGINE_URL?.includes('staging') ||
      env.CONVERSATION_ENGINE_URL?.includes('staging')) {
    return 'staging';
  }

  return 'production';
}

/**
 * Initialize and validate environment on startup
 */
export function initializeEnvironment(env: Env): EnvironmentConfig {
  const environmentType = detectEnvironment(env);
  const validation = validateEnvironment(env, environmentType);

  if (!validation.valid) {
    const errorMessage = `Environment validation failed:\n${validation.errors.join('\n')}`;
    console.error('❌ Environment Configuration Error:', errorMessage);
    throw new Error(errorMessage);
  }

  console.log(`✅ Environment validated successfully for ${environmentType}`);
  return getEnvironmentConfig(env);
}