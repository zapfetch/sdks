// Placeholder v1 example (TypeScript)
// Mirrors the older SDK usage. Replace with your API key before running.

// import ZapfetchApp from '@zapfetchdev/sdk';
import Zapfetch from '@zapfetchdev/sdk'

async function main() {
  const app = new Zapfetch({ apiKey: process.env.ZAPFETCH_API_KEY || 'fc-YOUR_API_KEY' });

  // Scrape a website (v1 style):
  const scrape = await app.v1.scrapeUrl('zapfetch.com');
  if ((scrape as any).success) console.log((scrape as any).markdown);

  // Crawl a website (v1 style):
  const crawl = await app.v1.crawlUrl('zapfetch.com', { excludePaths: ['blog/*'], limit: 3 });
  console.log(crawl);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});

