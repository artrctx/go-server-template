import { Hono } from "hono";
import { auth, type AuthType } from "../lib/auth/server.js";
import { cors } from "hono/cors";
import { APP_SCHEMA } from "../lib/auth/const.js";
import { env } from "../env.js";
import { db, user } from "../lib/auth/adapter.js";
import { eq, or } from "drizzle-orm";
import { tryCatch } from "../lib/util/trycatch.js";
import { HTTPException } from "hono/http-exception";

const router = new Hono<{ Bindings: AuthType }>({ strict: false });

const CORS_EXACT_MATCH = [env.API_URL];
const CORS_STARTS_WITH = [
  `${APP_SCHEMA}:\/\/`,
  ...(env.APP_ENV !== "production" ? ["exp://"] : []),
];

router.use(
  "*",
  cors({
    origin: (origin) => {
      return CORS_EXACT_MATCH.some((o) => o === origin) ||
        CORS_STARTS_WITH.some((o) => origin.startsWith(o))
        ? origin
        : null;
    },
    allowHeaders: ["Content-Type", "Authorization"],
    allowMethods: ["POST", "GET", "OPTIONS"],
    exposeHeaders: ["Content-Length"],
    maxAge: 600,
    credentials: true,
  }),
);

router.get("/auth/verify-availability", async (c) => {
  const username = c.req.query("username");
  const email = c.req.query("email");

  if (!username && !email) {
    c.status(400);
    return c.json({ message: "username or email is required as search query" });
  }

  const rows = await tryCatch(
    db
      .select({ email: user.email, username: user.username })
      .from(user)
      .where(
        username && email
          ? or(eq(user.username, username), eq(user.email, email))
          : username
            ? eq(user.username, username)
            : eq(user.email, email!),
      ),
  );

  if (rows.error) {
    c.status(500);
    return c.json({
      message: "failed with unknown error",
      cause: rows.error?.message,
    });
  }

  c.status(200);
  return c.json({
    usernameUnavailable: username
      ? rows.value.some((r) => r.username === username)
      : undefined,
    emailUnavailable: email
      ? rows.value.some((r) => r.email === email)
      : undefined,
  });
});

router.on(["POST", "GET"], "/auth/*", (c) => {
  return auth.handler(c.req.raw);
});

export { router as auth };
