import {
  createAccessControl,
  type Statements,
} from "better-auth/plugins/access";

// https://www.better-auth.com/docs/plugins/admin#access-control
export const statements = {
  deck: ["create", "read", "update", "delete"],
  user: ["invite", "ban"],
} as const satisfies Statements;

export const ac = createAccessControl(statements);

export const user = ac.newRole({
  deck: ["create", "read", "update", "delete"],
  user: ["invite"],
});

export const admin = ac.newRole({
  deck: ["create", "read", "update", "delete"],
  user: ["invite", "ban"],
});

export const moderator = ac.newRole({
  deck: ["create", "read", "update", "delete"],
  user: ["invite", "ban"],
});
