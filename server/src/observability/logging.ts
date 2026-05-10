import type { LoggerOptions } from 'pino';
import type { Config } from '../config.js';

export function buildLoggerOptions(cfg: Config): LoggerOptions {
  return {
    level: cfg.logLevel,
    redact: {
      paths: [
        'req.headers.authorization',
        'req.headers["x-api-key"]',
        'req.headers.cookie',
        'res.headers["set-cookie"]',
        '*.anthropicApiKey',
        '*.openaiApiKey',
        '*.apiKey',
      ],
      remove: true,
    },
    transport:
      cfg.nodeEnv === 'development'
        ? {
            target: 'pino-pretty',
            options: { colorize: true, translateTime: 'HH:MM:ss.l' },
          }
        : undefined,
  };
}
