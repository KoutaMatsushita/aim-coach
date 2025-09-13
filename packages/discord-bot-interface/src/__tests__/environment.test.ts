import { describe, expect, it, beforeEach } from '@jest/globals';
import { validateEnvironment, getEnvironmentConfig } from '../environment';
import type { Env } from '../types';

describe('Environment Configuration Management', () => {
  let mockEnv: Env;

  beforeEach(() => {
    mockEnv = {
      // Discord Bot settings
      DISCORD_TOKEN: 'test-discord-token',
      DISCORD_PUBLIC_KEY: 'test-public-key',
      DISCORD_APPLICATION_ID: 'test-app-id',

      // External services
      ANALYTICS_ENGINE_URL: 'https://analytics.example.com',
      CONVERSATION_ENGINE_URL: 'https://conversation.example.com',
      INTERNAL_API_TOKEN: 'test-internal-token',

      // Durable Objects
      USER_SESSION: {} as DurableObjectNamespace,

      // KV Storage
      BOT_CONFIG: {} as KVNamespace,
    };
  });

  describe('Environment Validation', () => {
    it('should validate complete environment configuration', () => {
      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    it('should fail validation when Discord token is missing', () => {
      delete (mockEnv as any).DISCORD_TOKEN;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('DISCORD_TOKEN is required');
    });

    it('should fail validation when Discord public key is missing', () => {
      delete (mockEnv as any).DISCORD_PUBLIC_KEY;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('DISCORD_PUBLIC_KEY is required');
    });

    it('should fail validation when Discord application ID is missing', () => {
      delete (mockEnv as any).DISCORD_APPLICATION_ID;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('DISCORD_APPLICATION_ID is required');
    });

    it('should fail validation when analytics engine URL is missing', () => {
      delete (mockEnv as any).ANALYTICS_ENGINE_URL;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('ANALYTICS_ENGINE_URL is required');
    });

    it('should fail validation when conversation engine URL is missing', () => {
      delete (mockEnv as any).CONVERSATION_ENGINE_URL;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('CONVERSATION_ENGINE_URL is required');
    });

    it('should fail validation when internal API token is missing', () => {
      delete (mockEnv as any).INTERNAL_API_TOKEN;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('INTERNAL_API_TOKEN is required');
    });

    it('should fail validation when Durable Objects binding is missing', () => {
      delete (mockEnv as any).USER_SESSION;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('USER_SESSION Durable Object binding is required');
    });

    it('should fail validation when KV namespace binding is missing', () => {
      delete (mockEnv as any).BOT_CONFIG;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('BOT_CONFIG KV namespace binding is required');
    });

    it('should collect multiple validation errors', () => {
      delete (mockEnv as any).DISCORD_TOKEN;
      delete (mockEnv as any).ANALYTICS_ENGINE_URL;
      delete (mockEnv as any).USER_SESSION;

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toHaveLength(3);
      expect(result.errors).toContain('DISCORD_TOKEN is required');
      expect(result.errors).toContain('ANALYTICS_ENGINE_URL is required');
      expect(result.errors).toContain('USER_SESSION Durable Object binding is required');
    });
  });

  describe('Environment Configuration Retrieval', () => {
    it('should return structured environment configuration', () => {
      const config = getEnvironmentConfig(mockEnv);

      expect(config).toEqual({
        discord: {
          token: 'test-discord-token',
          publicKey: 'test-public-key',
          applicationId: 'test-app-id',
        },
        services: {
          analyticsEngineUrl: 'https://analytics.example.com',
          conversationEngineUrl: 'https://conversation.example.com',
          internalApiToken: 'test-internal-token',
        },
        bindings: {
          userSession: mockEnv.USER_SESSION,
          botConfig: mockEnv.BOT_CONFIG,
        },
      });
    });

    it('should handle environment detection', () => {
      // Test for development environment indicators
      const config = getEnvironmentConfig(mockEnv);
      expect(config).toBeDefined();
    });
  });

  describe('Security Requirements', () => {
    it('should not log sensitive environment values', () => {
      const consoleSpy = jest.spyOn(console, 'log').mockImplementation();

      validateEnvironment(mockEnv);

      // Check that sensitive values are not logged
      const logCalls = consoleSpy.mock.calls.flat().join(' ');
      expect(logCalls).not.toContain('test-discord-token');
      expect(logCalls).not.toContain('test-public-key');
      expect(logCalls).not.toContain('test-internal-token');

      consoleSpy.mockRestore();
    });

    it('should validate URL formats for service endpoints', () => {
      mockEnv.ANALYTICS_ENGINE_URL = 'invalid-url';

      const result = validateEnvironment(mockEnv);
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('ANALYTICS_ENGINE_URL must be a valid URL');
    });

    it('should validate that URLs use HTTPS in production', () => {
      mockEnv.ANALYTICS_ENGINE_URL = 'http://analytics.example.com';

      const result = validateEnvironment(mockEnv, 'production');
      expect(result.valid).toBe(false);
      expect(result.errors).toContain('ANALYTICS_ENGINE_URL must use HTTPS in production');
    });
  });
});