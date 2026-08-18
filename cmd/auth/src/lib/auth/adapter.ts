import { drizzleAdapter } from "@better-auth/drizzle-adapter";
import { Pool } from "pg";
import { env } from "../../env.js";
import { drizzle } from "drizzle-orm/node-postgres";
import * as schema from "./schema.js";

export const db = drizzle(
  new Pool({
    connectionString: env.DATABASE_URL,
  }),
);

export const adapter = drizzleAdapter(db, {
  provider: "pg",
  schema,
});
