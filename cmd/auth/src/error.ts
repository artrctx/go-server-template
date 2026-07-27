import type { Hono } from "hono";
import type { AuthType } from "./lib/auth/server.js";
import type { BlankSchema } from "hono/types";
import { routePath } from "hono/route";
import { logger } from "./lib/logger/logger.js";
import { env } from "./env.js";

export function errorCaptureRegister(
  app: Hono<
    {
      Variables: AuthType;
    },
    BlankSchema,
    "/"
  >,
) {
  app.onError(async (err, c) => {
    const headers = c.req.header();
    const sanitizedHeaders = { ...headers };
    const sensitiveHeaders = [
      "authorization",
      "cookie",
      "x-api-key",
      "set-cookie",
    ];
    sensitiveHeaders.forEach((h) => delete sanitizedHeaders[h]);
    const userId = c.get("user")?.id || "anonymous";
    const userEmail = c.get("user")?.email || undefined;
    const errorCode =
      err && typeof err === "object" && "code" in err
        ? (err as Record<string, any>).code
        : undefined;

    logger.emit({
      severityText: "info",
      body: {
        // Request Context
        $current_url: c.req.url,
        request_path: c.req.path,
        request_method: c.req.method,
        request_headers: sanitizedHeaders,
        user_id: userId,

        // Route Context (Helps group errors by route rather than dynamic IDs)
        route_handler: routePath(c),

        // Error Metadata (PostHog handles stack traces, but explicit naming helps sorting)
        error_name: err.name || "Error",
        error_message: err.message,
        error_code: errorCode,

        // User Context (Optional, but highly recommended for filtering)
        user_email: userEmail,

        // Environment Context
        environment: env.APP_ENV || "development",
      },
    });

    // Standard internal server error response
    return c.text("Internal Server Error", 500);
  });
}
