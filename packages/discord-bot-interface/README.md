# Discord Bot Interface

AimCoach Discord Bot Interface - Cloudflare Workers implementation using discord-hono.

## 🚀 Features

- **Discord Slash Commands**: `/help`, `/analyze`, `/training`, `/status`
- **Automatic Signature Verification**: Handled by discord-hono
- **Type-Safe**: Full TypeScript support
- **Performance Optimized**: Built for Cloudflare Workers
- **Zero Dependencies**: discord-hono has zero runtime dependencies

## 📋 Commands

| Command | Description |
|---------|-------------|
| `/help` | Display available commands and usage |
| `/analyze` | Start aim score analysis workflow |
| `/training` | Get personalized training recommendations |
| `/status` | Show your current progress and recent scores |

## 🛠️ Development

### Prerequisites

- Node.js 18+
- Cloudflare Workers account
- Discord Application with Bot token

### Setup

1. Install dependencies:
```bash
npm install
```

2. Set up environment variables:
```bash
# Copy example configuration
cp wrangler.toml.example wrangler.toml

# Set secrets
wrangler secret put DISCORD_TOKEN
wrangler secret put DISCORD_PUBLIC_KEY
wrangler secret put DISCORD_APPLICATION_ID
```

3. Register Discord commands:
```bash
npm run register
```

4. Run development server:
```bash
npm run dev
```

### Testing

```bash
# Run all tests
npm test

# Run tests in watch mode
npm run test:watch

# Run linting
npm run lint

# Format code
npm run format
```

### Deployment

```bash
# Deploy to staging
wrangler deploy --env staging

# Deploy to production
wrangler deploy --env production
```

## 🏗️ Architecture

This implementation uses [discord-hono](https://discord-hono.luis.fun/) for:
- Automatic Discord signature verification
- Type-safe command handling
- Performance optimization for Cloudflare Workers
- Simplified Discord API integration

## 📚 Implementation Notes

- **TDD Approach**: All features implemented using Test-Driven Development
- **Zero Runtime Dependencies**: discord-hono has no runtime dependencies
- **Type Safety**: Full TypeScript support with strict type checking
- **Performance**: Optimized for Discord's 3-second timeout requirement

## 🔗 Related Services

- `analytics-engine`: Score analysis and feedback generation
- `conversation-engine`: Natural language processing for mentions
- `data-collector`: Local score data collection and synchronization