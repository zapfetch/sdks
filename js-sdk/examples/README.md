# Examples

Each file here is a standalone snippet that imports from `@zapfetchdev/sdk`.

To run them against a local checkout:

```bash
cd js-sdk
pnpm install
pnpm run build          # produces dist/ so @zapfetchdev/sdk resolves locally
npx tsx examples/example.ts
```

To run against the published package (once it's on npm):

```bash
mkdir my-zapfetch-demo && cd my-zapfetch-demo
npm init -y
npm install @zapfetchdev/sdk tsx
# copy any example_*.ts here, set ZAPFETCH_API_KEY, then:
npx tsx example.ts
```

## Files

| File | Demonstrates |
|---|---|
| `example.ts` / `example.js` | v2 scrape / extract / crawl / batch / search / map flows |
| `example_v1.ts` / `example_v1.js` | v1 legacy client (`ZapfetchAppV1`) |
| `example_pagination.ts` | v2 pagination helpers for long crawl / batch results |
| `example_watcher.ts` | v2 WebSocket watcher streaming job state |
