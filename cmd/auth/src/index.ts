import { serve } from "@hono/node-server";
import { Hono } from "hono";
import { auth } from "./routes/auth.js";
import type { AuthType } from "./lib/auth/server.js";
import { tryCatch } from "./lib/util/trycatch.js";
import { auth as authService } from "./lib/auth/server.js";
import { loggerMiddleware } from "./middleware/logger.js";
import { gracefulShutdown } from "./shutdown.js";
import { errorCaptureRegister } from "./error.js";
import { loggerSdk } from "./lib/logger/logger.js";
import { env } from "./env.js";

const app = new Hono<{ Variables: AuthType }>();

loggerSdk.start();

errorCaptureRegister(app);

app.use("*", loggerMiddleware);

app.get("/health", async (c) => {
  const res = await tryCatch(
    fetch(`${(await authService.$context).baseURL}/ok`),
  );
  if (res.error) {
    c.status(500);
    return c.json({ status: "down", cause: res.error });
  }
  return c.json({ status: res.value.ok ? "up" : "down" });
});

// https://better-auth.com/docs/integrations/hono
// https://hono.dev/examples/better-auth
const routers = [auth];

routers.forEach((r) => {
  app.basePath("/api").route("/", r);
});

const port = +(env.PORT ?? "3000");

const server = serve(
  {
    fetch: app.fetch,
    port: Number.isNaN(port) ? 3000 : port,
  },
  (info) => {
    console.log(`Server is running on http://localhost:${info.port}`);
  },
);

gracefulShutdown(server);
