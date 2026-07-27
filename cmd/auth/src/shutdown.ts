import type { ServerType } from "@hono/node-server";
import { logger, loggerSdk } from "./lib/logger/logger.js";

export function gracefulShutdown(server: ServerType) {
  const gracefulShutdown = async (signal: string) => {
    console.log(`Received ${signal}. Shutting down PostHog cleanly...`);
    try {
      server.close(async (err) => {
        if (!err) return;
        logger.emit({ severityText: "error", body: err.message });
      });
      await loggerSdk.shutdown();
    } catch (err) {
      console.error("Error during PostHog shutdown:", err);
      process.exit(1);
    }
    process.exit(0);
  };

  process.on("SIGTERM", () => gracefulShutdown("SIGTERM"));
  process.on("SIGINT", () => gracefulShutdown("SIGINT"));
}
