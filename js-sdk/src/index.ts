/**
 * Zapfetch JS/TS SDK — unified entrypoint.
 * - v2 by default on the top‑level client
 * - v1 available under `.v1` (feature‑frozen)
 * - Exports: `Zapfetch` (default), `ZapfetchClient` (v2), `ZapfetchAppV1` (v1), and v2 types
 */

/** Direct v2 client. */
export { ZapfetchClient } from "./v2/client";
/** Public v2 request/response types. */
export * from "./v2/types";
/** Watcher class and options for crawl/batch job monitoring. */
export { Watcher, type WatcherOptions } from "./v2/watcher";
/** Legacy v1 client (feature‑frozen). */
export { default as ZapfetchAppV1 } from "./v1";

import V1 from "./v1";
import { ZapfetchClient as V2, type ZapfetchClientOptions } from "./v2/client";
import type { ZapfetchAppConfig } from "./v1";

// Re-export v2 client options for convenience
export type { ZapfetchClientOptions } from "./v2/client";

/** Unified client: extends v2 and adds `.v1` for backward compatibility. */
export class Zapfetch extends V2 {
  /** Feature‑frozen v1 client (lazy). */
  private _v1?: V1;
  private _v1Opts: ZapfetchAppConfig;

  /** @param opts API credentials and base URL. */
  constructor(opts: ZapfetchClientOptions = {}) {
    super(opts);
    this._v1Opts = {
      apiKey: opts.apiKey,
      apiUrl: opts.apiUrl,
    };
  }

  /** Access the legacy v1 client (instantiated on first access). */
  get v1(): V1 {
    if (!this._v1) this._v1 = new V1(this._v1Opts);
    return this._v1;
  }
}

export default Zapfetch;

