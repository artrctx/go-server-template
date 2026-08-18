import { defineConfig } from "drizzle-kit";

export default defineConfig({
  dialect: "postgresql",
  schema: "./src/lib/auth/schema.ts",
  out: "./migrations",
  dbCredentials: {
    url:
      process.env.DIRECT_DATABASE_URL ||
      process.env.DATABASE_URL ||
      "postgres://postgres:password@localhost:5432/postgres",
  },
  migrations: {
    table: "auth-migrations",
    schema: "public",
  },
});
