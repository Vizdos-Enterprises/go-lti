import { TestInfo } from "@playwright/test";

type Part = "platform" | "lti";
export function Is3PC(test: TestInfo): boolean {
  return test.project.name.includes("3pc");
}

export function GetBaseURL(test: TestInfo, part: Part): string {
  return {
    platform: "https://platform.127-0-0-1.sslip.io:9999",
    lti: "https://tool.127.0.0.1.nip.io:9898",
  }[part];
}
