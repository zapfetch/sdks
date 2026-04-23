/**
 * Regression tests for HttpClient origin-injection logic.
 *
 * The SDK injects a default `origin` field into POST/PUT/PATCH bodies for
 * attribution, but must never overwrite an origin the caller has already set.
 */
import { HttpClient } from "../../../v2/utils/httpClient";

function setup() {
  const client = new HttpClient({ apiKey: "k", apiUrl: "https://api.zapfetch.com" });
  const captured: any[] = [];
  // Replace the internal axios instance with a stub that records the request
  // config. We go through bracket access to reach the private field.
  (client as any)["instance"] = {
    request: (cfg: any) => {
      captured.push(cfg);
      return Promise.resolve({ status: 200, data: {} });
    },
  };
  return { client, captured };
}

describe("HttpClient origin handling", () => {
  test("preserves caller-supplied origin on POST", async () => {
    const { client, captured } = setup();
    await client.post("/v2/scrape", { url: "https://ex.com", origin: "my-attribution-tag" });
    expect(captured[0].data.origin).toBe("my-attribution-tag");
  });

  test("injects SDK default origin when caller omits it", async () => {
    const { client, captured } = setup();
    await client.post("/v2/scrape", { url: "https://ex.com" });
    expect(captured[0].data.origin).toMatch(/^js-sdk@/);
  });

  test("treats empty string as 'not set' and uses SDK default", async () => {
    const { client, captured } = setup();
    await client.post("/v2/scrape", { url: "https://ex.com", origin: "" });
    expect(captured[0].data.origin).toMatch(/^js-sdk@/);
  });

  test("preserves MCP-style origin (parity with prior behavior for mcp markers)", async () => {
    const { client, captured } = setup();
    await client.post("/v2/scrape", { url: "https://ex.com", origin: "mcp-server" });
    expect(captured[0].data.origin).toBe("mcp-server");
  });

  test("preserves arbitrary non-mcp origin (regression fix)", async () => {
    const { client, captured } = setup();
    await client.post("/v2/scrape", { url: "https://ex.com", origin: "langchain-abc" });
    expect(captured[0].data.origin).toBe("langchain-abc");
  });
});
