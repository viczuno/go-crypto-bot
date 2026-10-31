import { describe, it, expect } from "vitest";
import { z } from "zod";
import { readErrorDetails } from "../lib/read-error";

describe("safe market read diagnostics", () => {
  it.each([
    ["42501", "permission"],
    ["42703", "sql_or_schema"],
    ["42P01", "sql_or_schema"],
    ["22P02", "database_value"],
    ["57014", "database"],
    ["08006", "database"],
  ])("preserves SQLSTATE %s without database payloads", (code, category) => {
    const error = Object.assign(new Error("secret connection string"), {
      code,
      detail: "sensitive row",
      query: "secret SQL",
      parameters: ["password"],
      table: "private",
      connectionString: "postgres://secret",
    });
    expect(readErrorDetails(error)).toEqual({ category, sqlstate: code });
  });
  it("redacts dynamic keys, values and Zod messages", () => {
    const result = z
      .object({
        analysis: z.object({ values: z.record(z.string(), z.number()) }),
      })
      .safeParse({ analysis: { values: { "sensitive-key": "password" } } });
    expect(result.success).toBe(false);
    if (!result.success) {
      const safe = readErrorDetails(result.error);
      expect(safe).toMatchObject({
        category: "validation",
        issues: [
          { code: "invalid_type", path: ["analysis", "values", "[redacted]"] },
        ],
      });
      expect(JSON.stringify(safe)).not.toMatch(/sensitive-key|password/);
    }
  });
  it("does not serialize unknown errors or untrusted codes", () => {
    expect(
      readErrorDetails(new Error("postgres://user:password@host")),
    ).toEqual({ category: "other" });
    expect(readErrorDetails({ code: "postgres://secret" })).toEqual({
      category: "other",
    });
  });
});
