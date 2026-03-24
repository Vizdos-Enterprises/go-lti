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

export async function ConfirmHeader(
  page: Page,
  header: "title-verify" | "title-verify-success" | "title-verify-failed",
) {
  await test.step(`confirm title is ${header}`, async () => {
    await expect(page.frameLocator("iframe").getByTestId(header)).toBeVisible();
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

    await test.step("open render page and wait for iframe auth init", async () => {
      await page.goto(`${platform}/render`);

      const frame = await getStartFrame(page);

      await WaitForLTIInit(frame);

      const newPagePromise = context.waitForEvent("page");
      const authSuccessPromise = waitForSpecificAuthMessage(frame, "auth_pass");

      await ConfirmHeader(page, "title-verify");

      await page.frameLocator("iframe").getByTestId("btn-continue").click();

      const popup = await newPagePromise;
      await popup.waitForLoadState();

      await authSuccessPromise;
      await expect(popup).toHaveURL(`${lti}/lti/app/`);

      await ConfirmHeader(page, "title-verify-success");
    });

    await test.step("clicking Open Activity works once loaded", async () => {
      await ConfirmHeader(page, "title-verify-success");
      const newPagePromise = context.waitForEvent("page");

      await page.frameLocator("iframe").getByTestId("btn-continue").click();
      const popup = await newPagePromise;
      await popup.waitForLoadState();

      await expect(popup).toHaveURL(`${lti}/lti/app/`);
    });

    await test.step("updates original launcher when token is expired / missing", async () => {
      await ConfirmHeader(page, "title-verify-success");
      await page.context().clearCookies();
      const newPagePromise = context.waitForEvent("page");

      await page.frameLocator("iframe").getByTestId("btn-continue").click();
      const popup = await newPagePromise;
      await popup.waitForLoadState();

      await expect(popup).toHaveURL(`${lti}/lti/auth/error?err=missing+token`);
      await expect(popup.getByTestId("error-title")).toHaveText(
        "Session Missing",
      );

      await ConfirmHeader(page, "title-verify-failed");
    });
  });

  test("times out if popup is closed before confirmation", async ({
    page,
    context,
  }, testInfo) => {
    const platform = GetBaseURL(testInfo, "platform");

    await page.goto(`${platform}/render`);

    const frame = await getStartFrame(page);
    await WaitForLTIInit(frame);

    await ConfirmHeader(page, "title-verify");

    const popupPromise = context.waitForEvent("page");

    await context.route("**/lti/auth/exchange", async (route) => {
      await new Promise((r) => setTimeout(r, 5000)); // 5s delay
      await route.continue();
    });

    await page.frameLocator("iframe").getByTestId("btn-continue").click();

    const popup = await popupPromise;
    await popup.waitForLoadState();

    await popup.close();

    // Now wait for timeout UI (should be < 10s based on your 8s timer)
    const start = Date.now();

    await expect(
      page.frameLocator("iframe").getByTestId("title-verify-failed"),
    ).toBeVisible({ timeout: 10000 });

    await expect(
      page.frameLocator("iframe").getByTestId("title-verify-failed"),
    ).toHaveText("Confirmation took too long.");

    const duration = Date.now() - start;

    // sanity: make sure it didn't take forever
    expect(duration).toBeLessThan(10000);

    // assert timeout-specific UI
    await expect(
      page.frameLocator("iframe").getByTestId("error-code"),
    ).toHaveText("Confirmation Timeout");
  });
});
