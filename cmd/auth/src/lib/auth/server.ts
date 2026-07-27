import { betterAuth } from "better-auth";
import {
  admin as adminConfig,
  jwt,
  username,
  emailOTP,
} from "better-auth/plugins";
import { ac, admin, moderator, user } from "./permissions.js";
import { APP_SCHEMA } from "./const.js";
import { expo } from "@better-auth/expo";
import { env } from "../../env.js";
import { logger } from "../logger/logger.js";
import type { AnyValue } from "@opentelemetry/api-logs";
import { adapter } from "./adapter.js";
import { resend } from "../server/email/email.js";
import { AUTH_EMAIL_CONTENT } from "./emails.js";

export const auth = betterAuth({
  database: adapter,
  baseURL: env.BETTER_AUTH_URL,
  trustedOrigins: [
    `${APP_SCHEMA}:\/\/`,
    `${APP_SCHEMA}:\/\/*`,
    env.API_URL,
    ...(env.APP_ENV !== "production"
      ? [
          "exp://", // Trust all Expo URLs (prefix matching)
          "exp://**", // Trust all Expo URLs (wildcard matching)
          "exp://192.168.*.*:*/**", // Trust 192.168.x.x IP range with any port and path
        ]
      : []),
  ],
  logger: {
    disabled: false,
    disableColors: false,
    level: env.APP_ENV === "production" ? "info" : "debug",
    log: (level, message, ...args) => {
      logger.emit({
        eventName: "auth_logger",
        severityText: level,
        body: {
          message,
          ...args,
        },
      });
    },
  },
  onAPIError: {
    onError: (error, ctx) => {
      // Custom error handling
      console.error("Auth error:", error);
      logger.emit({
        eventName: "auth_error",
        severityText: "error",
        body: {
          error: error as AnyValue,
          user: ctx.session?.user,
        },
      });
    },
  },
  account: {
    accountLinking: {
      enabled: true,
      trustedProviders: ["google"],
    },
  },
  plugins: [
    jwt(),
    username(),
    expo(),
    adminConfig({
      ac,
      roles: {
        user,
        moderator,
        admin,
      },
    }),
    emailOTP({
      otpLength: 6,
      expiresIn: 300,
      overrideDefaultEmailVerification: true,
      disableSignUp: true,
      async sendVerificationOTP({ email, otp, type }) {
        const { data, error } = await resend.emails.send({
          from: `Shuffle <${env.APP_EMAIL}>`,
          to: email,
          ...AUTH_EMAIL_CONTENT[type](otp),
        });

        if (error)
          logger.emit({
            eventName: "email_verification_send_error",
            severityText: "error",
            body: {
              type,
              error: error as AnyValue,
              toEmail: email,
            },
          });
        else
          logger.emit({
            eventName: "email_verification_sent",
            severityText: "info",
            body: {
              type,
              emailId: data.id,
              toEmail: email,
            },
          });
      },
    }),
  ],
  emailAndPassword: {
    enabled: true,
    autoSignIn: true,
  },
  emailVerification: {
    autoSignInAfterVerification: true,
    sendOnSignUp: true,
  },
  socialProviders: {
    google: {
      clientId: env.GOOGLE_CLIENT_ID as string,
      clientSecret: env.GOOGLE_CLIENT_SECRET as string,
    },
  },
  session: {
    cookieCache: {
      enabled: true,
      maxAge: 5 * 60, // Cache expiration time in seconds (e.g., 5 minutes)
      strategy: "jwt", // Changes the cookie format to standard, signed JWT
    },
  },
  advanced: {
    cookiePrefix: APP_SCHEMA,
    disableOriginCheck: env.APP_ENV === "development",
    disableCSRFCheck: env.APP_ENV === "development",
    crossSubDomainCookies: {
      enabled: true,
    },
  },
});

export type AuthType = {
  user: typeof auth.$Infer.Session.user | null;
  session: typeof auth.$Infer.Session.session | null;
};
