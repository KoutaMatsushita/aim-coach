import { jest } from '@jest/globals';

// Mock Cloudflare Workers environment
global.fetch = jest.fn();

// Mock environment variables
const mockEnv = {
  DISCORD_TOKEN: 'mock-discord-token',
  DISCORD_PUBLIC_KEY: 'mock-public-key',
  DISCORD_APPLICATION_ID: 'mock-app-id',
  ANALYTICS_ENGINE_URL: 'http://localhost:3001',
  CONVERSATION_ENGINE_URL: 'http://localhost:3002',
  INTERNAL_API_TOKEN: 'mock-internal-token',
  USER_SESSION: {} as DurableObjectNamespace,
  BOT_CONFIG: {} as KVNamespace,
};

// Make mock env available globally for tests
(global as any).mockEnv = mockEnv;

// Setup console methods for testing
global.console = {
  ...console,
  warn: jest.fn(),
  error: jest.fn(),
};