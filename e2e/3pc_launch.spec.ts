import { test, expect, Page, Frame } from "@playwright/test";
import { GetBaseURL, Is3PC } from "./helper.ts";

async function getStartFrame(page: Page): Promise<Frame> {
  const iframeHandle = await page.locator("iframe").elementHandle();
  if (!iframeHandle) throw new Error("iframe not found");

  const frame = await iframeHandle.contentFrame();
  if (!frame) throw new Error("iframe contentFrame not available");

  return frame;
}

export async function WaitForLTIInit(frame: Frame) {
  await test.step("wait for LTI init", async () => {
    await frame.waitForFunction(() => !!(window as any).__ltiAuthInitReady);
    await frame.waitForTimeout(500);
  });
}

async function waitForSpecificAuthMessage(
  frame: Frame,
  expectedType: "auth_pass" | "auth_failure" | "exchange_burned",
) {
  return frame.evaluate((expectedType) => {
    return new Promise((resolve) => {
      function handler(e: MessageEvent) {
        if (e.origin !== window.location.origin) return;
        if (e.data?.type !== expectedType) return;

        window.removeEventListener("message", handler);
        resolve(e.data);
      }

      window.addEventListener("message", handler);
    });
  }, expectedType);
}

test.describe("3pc tests", () => {
  test.beforeEach(async ({}, testInfo) => {
    test.skip(!Is3PC(testInfo), "3pc-only suite");
  });

  test("fallback works", async ({ page, context }, testInfo) => {
    const platform = GetBaseURL(testInfo, "platform");
    const lti = GetBaseURL(testInfo, "lti");

    await page.goto(`${platform}/render`);

    const frame = await getStartFrame(page);

    await WaitForLTIInit(frame);

    const newPagePromise = context.waitForEvent("page");
    const authSuccessPromise = waitForSpecificAuthMessage(frame, "auth_pass");

    await page.frameLocator("iframe").getByTestId("btn-continue").click();

    const popup = await newPagePromise;
    await popup.waitForLoadState();

    await authSuccessPromise;
    await expect(popup).toHaveURL(`${lti}/lti/app/`);
  });
});
