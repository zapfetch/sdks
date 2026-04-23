import ZapfetchApp from '@zapfetchdev/sdk';

// Placeholder v1 example (JavaScript)
// Mirrors the older SDK usage. Replace with your API key before running.

async function main() {
  const app = new Zapfetch({ apiKey: process.env.ZAPFETCH_API_KEY || 'fc-YOUR_API_KEY' });

  const scrape = await app.v1.scrapeUrl('zapfetch.com');
  if (scrape && scrape.success) console.log(scrape.markdown);

  const crawl = await app.v1.crawlUrl('zapfetch.com', { excludePaths: ['blog/*'], limit: 3 });
  console.log(crawl);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});