import { createEnv } from "@t3-oss/env-core";
import * as z from "zod";

export const env = createEnv({
  server: {
    DATABASE_URL: z.url(),
    POSTHOG_TOKEN: z.string().nonempty(),
    BETTER_AUTH_SECRET: z.string().nonempty(),
    BETTER_AUTH_URL: z.string().nonempty(),
    PORT: z.string(),
    API_URL: z.string().nonempty(),
    APP_ENV: z.string().nonempty(),
    APP_EMAIL: z.string().nonempty(),
    GOOGLE_CLIENT_ID: z.string().nonempty(),
    GOOGLE_CLIENT_SECRET: z.string().nonempty(),
    RESEND_API_KEY: z.string().nonempty(),
  },
  runtimeEnv: process.env,
});
