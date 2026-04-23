// Placeholder v2 example (TypeScript)
// Minimal usage of new ZapfetchClient. Replace with your API key before running.

// import Zapfetch from '@zapfetchdev/sdk';
import Zapfetch from '@zapfetchdev/sdk';

const run = async () => {
  const apiKey = (globalThis as any).process?.env?.ZAPFETCH_API_KEY || 'fc-YOUR_API_KEY';
  const apiUrl = (globalThis as any).process?.env?.ZAPFETCH_API_URL || 'https://api.zapfetch.com';
  const client = new Zapfetch({ apiKey, apiUrl });

  const doc = await client.scrape('https://docs.zapfetch.com', { formats: ['markdown'] });
  console.log('scrape:', !!doc.markdown);

  const crawl = await client.crawl('https://docs.zapfetch.com', { limit: 3, pollInterval: 1, timeout: 120 });
  console.log('crawl:', crawl.status, crawl.completed, '/', crawl.total);

  const batch = await client.batchScrape([
    'https://docs.zapfetch.com',
    'https://zapfetch.com',
  ], { options: { formats: ['markdown'] }, pollInterval: 1, timeout: 120 });
  console.log('batch:', batch.status, batch.completed, '/', batch.total);

  const search = await client.search('What is the capital of France?', { limit: 5 });
  console.log('search web results:', (search.web || []).length);

  const map = await client.map('https://zapfetch.com');
  console.log('map links:', map.links.length);
};

run().catch((e) => {
  // eslint-disable-next-line no-console
  console.error(e);
  (globalThis as any).process?.exit?.(1);
});

// old stuff:

import ZapfetchApp, { CrawlStatusResponse, ErrorResponse } from '@zapfetchdev/sdk';

const app = new ZapfetchApp({apiKey: "fc-YOUR_API_KEY"});

const main = async () => {

  // Scrape a website:
  const scrapeResult = await app.scrapeUrl('zapfetch.com');

  if (scrapeResult.success) {
    console.log(scrapeResult.markdown)
  }

  // Crawl a website:
  const crawlResult = await app.crawlUrl('zapfetch.com', { excludePaths: ['blog/*'], limit: 5});
  console.log(crawlResult);

  // Asynchronously crawl a website:
  const asyncCrawlResult = await app.asyncCrawlUrl('zapfetch.com', { excludePaths: ['blog/*'], limit: 5});
  
  if (asyncCrawlResult.success) {
    const id = asyncCrawlResult.id;
    console.log(id);

    let checkStatus: CrawlStatusResponse | ErrorResponse;
    if (asyncCrawlResult.success) {
      while (true) {
        checkStatus = await app.checkCrawlStatus(id);
        if (checkStatus.success && checkStatus.status === 'completed') {
          break;
        }
        await new Promise(resolve => setTimeout(resolve, 1000)); // wait 1 second
      }

      if (checkStatus.success && checkStatus.data) {
        console.log(checkStatus.data[0].markdown);
      }
    }
  }

  // Map a website:
  const mapResult = await app.mapUrl('https://zapfetch.com');
  console.log(mapResult)

  // // Extract information from a website using LLM:
  // const extractSchema = z.object({
  //   title: z.string(),
  //   description: z.string(),
  //   links: z.array(z.string())
  // });

  // const extractResult = await app.extractUrls(['https://zapfetch.com'], {
  //   prompt: "Extract the title, description, and links from the website",
  //   schema: extractSchema
  // });
  // console.log(extractResult);

  // Crawl a website with WebSockets:
  const watch = await app.crawlUrlAndWatch('zapfetch.com', { excludePaths: ['blog/*'], limit: 5});

  watch.addEventListener("document", doc => {
    console.log("DOC", doc.detail);
  });

  watch.addEventListener("error", err => {
    console.error("ERR", err.detail.error);
  });

  watch.addEventListener("done", state => {
    console.log("DONE", state.detail.status);
  });
}

main()