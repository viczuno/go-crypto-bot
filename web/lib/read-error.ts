import { ZodError } from "zod";

const safeFields = new Set([
  "id",
  "name",
  "symbol",
  "active",
  "price",
  "cap",
  "volume",
  "change",
  "updated",
  "history",
  "time",
  "analysis",
  "day",
  "samples",
  "signal",
  "strength",
  "coverage",
  "indicators",
  "value",
  "values",
  "required",
  "reason",
]);
export type ReadStage =
  | "connect"
  | "assets_query"
  | "runs_query"
  | "asset_validation"
  | "run_mapping"
  | "history_query"
  | "analytics_query"
  | "detail_validation";

// Never serialize error messages, stacks, queries, parameters, or driver detail:
// PostgreSQL and validation messages may contain connection details or row values.
export function readErrorDetails(error: unknown) {
  if (error instanceof ZodError) {
    return {
      category: "validation",
      issues: error.issues.slice(0, 20).map((issue) => ({
        code: issue.code,
        path: issue.path
          .slice(0, 12)
          .map((part) =>
            typeof part === "number"
              ? part
              : typeof part === "string" && safeFields.has(part)
                ? part
                : "[redacted]",
          ),
      })),
    };
  }
  const code =
    typeof error === "object" && error !== null && "code" in error
      ? error.code
      : undefined;
  if (
    typeof code === "string" &&
    /^(?:[0-9]{2}|XX|P0|HV|F0)[A-Z0-9]{3}$/.test(code)
  ) {
    return {
      category:
        code === "42501"
          ? "permission"
          : code.startsWith("42")
            ? "sql_or_schema"
            : code.startsWith("22")
              ? "database_value"
              : "database",
      sqlstate: code,
    };
  }
  return { category: "other" };
}
export function logReadError(
  event: "market_read_failed" | "market_detail_read_failed",
  stage: ReadStage,
  error: unknown,
) {
  console.error(JSON.stringify({ event, stage, ...readErrorDetails(error) }));
}
