package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.Document;
import com.zapfetch.models.ScrapeOptions;

import java.util.List;

/**
 * Scrape a single URL and print its markdown.
 *
 * Usage:
 *   export ZAPFETCH_API_KEY=zf-...
 *   javac -cp <zapfetch-sdk.jar> ScrapeExample.java
 *   java  -cp <zapfetch-sdk.jar>:. com.zapfetch.examples.ScrapeExample https://example.com
 */
public final class ScrapeExample {
    private ScrapeExample() {}

    public static void main(String[] args) {
        String target = args.length > 0 ? args[0] : "https://example.com";

        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        Document doc = client.scrape(target, ScrapeOptions.builder()
            .formats(List.of("markdown"))
            .onlyMainContent(true)
            .build());

        System.out.println("--- " + target + " ---");
        System.out.println(doc.getMarkdown());
    }
}
