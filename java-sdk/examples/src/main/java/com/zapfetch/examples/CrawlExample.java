package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.CrawlJob;
import com.zapfetch.models.CrawlOptions;
import com.zapfetch.models.Document;
import com.zapfetch.models.ScrapeOptions;

import java.util.List;

/**
 * Crawl a site with depth + page-count limits, then summarise each result.
 *
 * Usage:
 *   ZAPFETCH_API_KEY=zf-... java ... com.zapfetch.examples.CrawlExample https://example.com
 */
public final class CrawlExample {
    private CrawlExample() {}

    public static void main(String[] args) {
        String target = args.length > 0 ? args[0] : "https://example.com";

        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        CrawlJob job = client.crawl(target, CrawlOptions.builder()
            .limit(10)
            .maxDiscoveryDepth(2)
            .scrapeOptions(ScrapeOptions.builder()
                .formats(List.of("markdown"))
                .onlyMainContent(true)
                .build())
            .build());

        System.out.printf("crawl status=%s completed=%d/%d%n",
            job.getStatus(), job.getCompleted(), job.getTotal());

        int i = 1;
        for (Document doc : job.getData()) {
            String title = doc.getMetadata() != null
                ? String.valueOf(doc.getMetadata().getOrDefault("title", ""))
                : "";
            int len = doc.getMarkdown() == null ? 0 : doc.getMarkdown().length();
            System.out.printf("[%2d] %6d chars  %s%n", i++, len, title);
        }
    }
}
