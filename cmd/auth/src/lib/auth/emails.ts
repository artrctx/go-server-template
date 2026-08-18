import { type EmailOTPOptions } from "better-auth/plugins";

export type AuthEmailContent = Record<
  Parameters<EmailOTPOptions["sendVerificationOTP"]>[0]["type"],
  (otp: string) => { subject: string; html: string; text: string }
>;

export const AUTH_EMAIL_CONTENT = {
  "sign-in": (otp) => ({
    subject: "Sign in request from templ",
    html: `<div><p>Sign in verification otp:</p><p>${otp}</p></div>`,
    text: `Sign in verification otp: ${otp}`,
  }),
  "email-verification": (otp) => ({
    subject: "Email verification otp from templ",
    html: `<div><p>Email verification otp:</p><p>${otp}</p></div>`,
    text: `Email verification otp: ${otp}`,
  }),
  "forget-password": (otp) => ({
    subject: "Password reset otp from templ",
    html: `<div><p>Forgot password otp:</p><p>${otp}</p></div>`,
    text: `Forgot password verification otp: ${otp}`,
  }),
  "change-email": (otp) => ({
    subject: "Change email request otp from templ",
    html: `<div><p>Change email request otp:</p><p>${otp}</p></div>`,
    text: `Change email request otp: ${otp}`,
  }),
} as const satisfies AuthEmailContent;
