import { buildApp } from './app.js';
import { loadConfig } from './config.js';

async function main(): Promise<void> {
  const cfg = loadConfig();
  const app = await buildApp(cfg);

  try {
    await app.listen({ port: cfg.port, host: '0.0.0.0' });
  } catch (err) {
    app.log.error(err, 'Failed to start server');
    process.exit(1);
  }
}

main().catch((err) => {
  console.error('Fatal startup error:', err);
  process.exit(1);
});
