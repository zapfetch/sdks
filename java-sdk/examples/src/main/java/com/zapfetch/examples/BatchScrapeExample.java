package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.BatchScrapeJob;
import com.zapfetch.models.BatchScrapeOptions;
import com.zapfetch.models.Document;
import com.zapfetch.models.ScrapeOptions;

import java.util.List;

/**
 * Batch-scrape a list of URLs and print titles.
 *
 * Usage:
 *   ZAPFETCH_API_KEY=zf-... java ... com.zapfetch.examples.BatchScrapeExample
 */
public final class BatchScrapeExample {
    private BatchScrapeExample() {}

    public static void main(String[] args) {
        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        List<String> urls = List.of(
            "https://example.com",
            "https://example.org",
            "https://example.net"
        );

        BatchScrapeJob job = client.batchScrape(urls, BatchScrapeOptions.builder()
            .options(ScrapeOptions.builder()
                .formats(List.of("markdown"))
                .onlyMainContent(true)
                .build())
            .maxConcurrency(3)
            .build());

        System.out.printf("batch status=%s completed=%d/%d%n",
            job.getStatus(), job.getCompleted(), job.getTotal());

        int i = 1;
        for (Document doc : job.getData()) {
            var meta = doc.getMetadata();
            String title = meta != null ? String.valueOf(meta.getOrDefault("title", "")) : "";
            String source = meta != null ? String.valueOf(meta.getOrDefault("sourceURL", "")) : "";
            System.out.printf("[%2d] %s%n     %s%n", i++, title, source);
        }
    }
}
