import { createMiddleware } from "hono/factory";
import { routePath } from "hono/route";
import { getConnInfo } from "@hono/node-server/conninfo";
import { logger } from "../lib/logger/logger.js";
import { tryCatch } from "../lib/util/trycatch.js";

export const loggerMiddleware = createMiddleware(async (c, next) => {
  if (
    c.req.path.startsWith("/favicon.ico") ||
    c.req.path.startsWith("/static")
  ) {
    return await next();
  }

  const startTime = Date.now();

  await next();

  const user = c.get("user");
  const distinctId = user?.id || user?.email || "anonymous_server_request";

  logger.emit({
    severityText: c.res.status >= 400 ? "error" : "info",
    body: {
      distinctId: distinctId,
      event: "auth_server_request",
      properties: {
        $current_url: c.req.url,
        ip: getConnInfo(c).remote.address,
        path: c.req.path,
        method: c.req.method,
        status: c.res.status, // Capture if it was a 200, 404, 500, etc.
        duration_ms: Date.now() - startTime,
        resBody: (await tryCatch(c.res.clone().text())).value,
        route: routePath(c),
        ...(user && {
          user_email: user.email,
          user_role: user.role,
        }),
      },
    },
  });
});
