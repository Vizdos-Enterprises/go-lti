import { test, expect, Frame, Page } from "@playwright/test";
import { GetBaseURL, Is3PC } from "./helper.ts";

async function getFrame(page: Page): Promise<Frame> {
  const handle = await page.locator("iframe").elementHandle();
  if (!handle) throw new Error("iframe not found");

  const frame = await handle.contentFrame();
  if (!frame) throw new Error("iframe contentFrame not available");

  return frame;
}

test("launch works", async ({ page }, testInfo) => {
  const platform = GetBaseURL(testInfo, "platform");
  const lti = GetBaseURL(testInfo, "lti");

  await page.goto(`${platform}/render`);

  const frame = await getFrame(page);

  if (Is3PC(testInfo)) {
    await expect.poll(() => frame.url()).toMatch(/\/lti\/auth\/verify/);
  } else {
    await expect.poll(() => frame.url()).toBe(`${lti}/lti/app/`);
  }
});
