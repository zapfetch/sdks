package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.SearchData;
import com.zapfetch.models.SearchOptions;

import java.util.List;
import java.util.Map;

/**
 * Search the web and print the top results.
 *
 * Usage:
 *   ZAPFETCH_API_KEY=zf-... java ... com.zapfetch.examples.SearchExample "web scraping tools"
 */
public final class SearchExample {
    private SearchExample() {}

    public static void main(String[] args) {
        String query = args.length > 0 ? String.join(" ", args) : "web scraping tools";

        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        SearchData results = client.search(query, SearchOptions.builder()
            .limit(5)
            .build());

        System.out.printf("query=%s%n", query);
        List<Map<String, Object>> web = results.getWeb();
        if (web != null) {
            int i = 1;
            for (Map<String, Object> r : web) {
                System.out.printf("[%2d] %s%n     %s%n",
                    i++,
                    r.getOrDefault("title", ""),
                    r.getOrDefault("url", ""));
            }
        }
    }
}
