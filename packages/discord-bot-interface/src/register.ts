import { Command, register } from 'discord-hono';

/**
 * Register Discord slash commands with Discord API
 * Run this script after deploying the worker to register commands
 */

const commands = [
  new Command('help', 'AimCoachの利用可能なコマンドを表示します'),
  new Command('analyze', 'あなたのエイムスコアを分析します'),
  new Command('training', 'パーソナライズされたトレーニング推奨を生成します'),
  new Command('status', 'あなたの現在の進捗状況を表示します'),
];

/**
 * Register commands with Discord
 * Make sure to set environment variables:
 * - DISCORD_APPLICATION_ID
 * - DISCORD_TOKEN
 * - DISCORD_TEST_GUILD_ID (optional, for testing in specific guild)
 */
async function registerCommands(): Promise<void> {
  try {
    console.log('🚀 Registering Discord commands...');

    await register(
      commands,
      process.env.DISCORD_APPLICATION_ID!,
      process.env.DISCORD_TOKEN!,
      // process.env.DISCORD_TEST_GUILD_ID // Uncomment for guild-specific registration
    );

    console.log('✅ Commands registered successfully!');
    console.log(`📋 Registered ${commands.length} commands:`);
    commands.forEach(cmd => {
      console.log(`   /${cmd.name} - ${cmd.description}`);
    });

  } catch (error) {
    console.error('❌ Failed to register commands:', error);
    process.exit(1);
  }
}

// Run if this file is executed directly
if (require.main === module) {
  registerCommands();
}