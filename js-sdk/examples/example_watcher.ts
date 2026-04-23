/*
 Advanced watcher example using the v2 ZapfetchClient.

 Run with:
   node --env-file=.env -r esbuild-register apps/js-sdk/zapfetch/src/examples/watcher.ts
 or compile with your bundler, ensuring ZAPFETCH_API_KEY is set.
*/

import { ZapfetchClient } from "@zapfetchdev/sdk";

async function main() {
  const apiKey = process.env.ZAPFETCH_API_KEY || "fc-YOUR_API_KEY";
  const client = new ZapfetchClient({ apiKey });

  // Start a crawl and attach a watcher for real-time updates
  const start = await client.startCrawl("https://example.com", { limit: 5 });
  console.log("Started crawl:", start.id);

  const watcher = client.watcher(start.id, { kind: "crawl", pollInterval: 2, timeout: 120 });

  watcher.on("document", (doc) => {
    console.log("document:", (doc as any).url || (doc as any).metadata?.sourceURL || "<no url>");
  });

  watcher.on("snapshot", (snap) => {
    console.log(`status: ${snap.status} (${snap.completed}/${snap.total})`);
  });

  watcher.on("done", (finalState) => {
    console.log("done:", finalState.status, "docs:", finalState.data?.length ?? 0);
  });

  await watcher.start();
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});

