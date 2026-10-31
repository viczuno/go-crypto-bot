import { z } from "zod";
export const pointSchema = z.object({
  time: z.string(),
  price: z.number().positive(),
});
export type Point = z.infer<typeof pointSchema>;
export const analysisSchema = z.object({
  day: z.string().optional(),
  samples: z.number(),
  signal: z.string(),
  strength: z.number(),
  coverage: z.number(),
  indicators: z.array(
    z.object({
      name: z.string(),
      value: z.number().nullable(),
      signal: z.string(),
      strength: z.number(),
      values: z.record(z.string(), z.number()),
      required: z.number(),
      reason: z.string().optional(),
    }),
  ),
});
export type Analysis = z.infer<typeof analysisSchema>;
export const assetSchema = z.object({
  id: z.string(),
  name: z.string(),
  symbol: z.string(),
  active: z.boolean(),
  price: z.number().nullable(),
  cap: z.number().nullable(),
  volume: z.number().nullable(),
  change: z.number().nullable(),
  updated: z.string().nullable(),
  history: z.array(pointSchema),
  analysis: analysisSchema.nullable(),
});
export type Asset = z.infer<typeof assetSchema>;
export type Snapshot = {
  assets: Asset[];
  status: "ready" | "empty" | "unavailable";
  run: {
    status: string;
    finished: string | null;
    received: number;
    expected: number;
  } | null;
  asOf: string;
};
export const querySchema = z.object({
  q: z.string().max(80).default(""),
  sort: z.enum(["cap", "price", "change", "volume", "name"]).default("cap"),
  direction: z.enum(["asc", "desc"]).default("desc"),
  page: z.coerce.number().int().min(1).max(10000).default(1),
  limit: z.coerce.number().int().min(1).max(100).default(25),
});
export const idSchema = z.string().regex(/^[a-z0-9][a-z0-9-]{0,99}$/);
