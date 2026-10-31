import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
test("dashboard, watchlist, filters and comparison", async ({ page }) => {
  const errors: string[] = [];
  page.on("pageerror", (e) => errors.push(e.message));
  await page.goto("/");
  await expect(
    page.getByRole("heading", { name: "Market overview" }),
  ).toBeVisible();
  await expect(page.getByText("Market snapshot")).toBeVisible();
  await expect(page.locator("canvas").first()).toBeVisible();
  await expect(page.locator(".recharts-surface").first()).toBeVisible();
  await page.screenshot({
    path: `test-results/dashboard-${test.info().project.name}.png`,
    fullPage: true,
  });
  await page.goto("/market");
  await page.getByRole("textbox", { name: "Search assets" }).fill("bitcoin");
  const bitcoin = page
    .locator("tbody tr")
    .filter({ has: page.locator('a[href="/coins/bitcoin"]') });
  await expect(bitcoin).toHaveCount(1);
  await bitcoin.getByRole("button", { name: "Add to watchlist" }).click();
  await page.goto("/watchlist");
  await expect(page.locator("tbody tr")).toHaveCount(1);
  await page.reload();
  await expect(page.locator("tbody tr")).toHaveCount(1);
  await page.goto("/coins/bitcoin");
  await expect(
    page.getByRole("heading", { name: "Technical analysis" }),
  ).toBeVisible();
  await expect(
    page.getByRole("combobox", { name: "Price overlay" }),
  ).toHaveCount(1);
  await page
    .getByRole("combobox", { name: "Price overlay" })
    .selectOption("BollingerBands");
  await page.getByRole("button", { name: "30D", exact: true }).click();
  await page.getByRole("button", { name: "MACD", exact: true }).click();
  await page.goto("/compare");
  await expect(page.getByText("BTC", { exact: true }).first()).toBeVisible();
  await expect(page.locator("canvas").first()).toBeVisible();
  await page.getByRole("button", { name: "Toggle color theme" }).click();
  await page.screenshot({
    path: `test-results/compare-light-${test.info().project.name}.png`,
    fullPage: true,
  });
  expect(errors).toEqual([]);
});
test("accessibility and bounded API", async ({ page, request }) => {
  await page.goto("/market");
  const result = await new AxeBuilder({ page })
    .withTags(["wcag2a", "wcag2aa"])
    .analyze();
  expect(result.violations).toEqual([]);
  const response = await request.get("/api/market?limit=10000");
  expect(response.status()).toBe(400);
  const good = await request.get("/api/market?limit=2");
  expect(good.status()).toBe(200);
  expect((await good.json()).assets).toHaveLength(2);
  await page.goto("/coins/does-not-exist");
  await expect(
    page.getByRole("heading", { name: "Asset not found" }),
  ).toBeVisible();
});
