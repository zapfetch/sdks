/**
 * Minimal unit test for v2 scrape (no mocking; sanity check payload path)
 */
import { ZapfetchClient } from "../../../v2/client";

describe("v2.scrape unit", () => {
  test("constructor requires apiKey", () => {
    expect(() => new ZapfetchClient({ apiKey: "", apiUrl: "https://api.zapfetch.com" })).toThrow();
  });
});

