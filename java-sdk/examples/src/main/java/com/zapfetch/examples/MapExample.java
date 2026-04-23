package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.MapData;
import com.zapfetch.models.MapOptions;

import java.util.List;
import java.util.Map;

/**
 * Discover URLs on a site using the map endpoint.
 *
 * Usage:
 *   ZAPFETCH_API_KEY=zf-... java ... com.zapfetch.examples.MapExample https://example.com pricing
 */
public final class MapExample {
    private MapExample() {}

    public static void main(String[] args) {
        String target = args.length > 0 ? args[0] : "https://example.com";
        String search = args.length > 1 ? args[1] : null;

        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        MapOptions.Builder opts = MapOptions.builder()
            .limit(50)
            .includeSubdomains(true);
        if (search != null) {
            opts.search(search);
        }

        MapData data = client.map(target, opts.build());
        List<Map<String, Object>> links = data.getLinks();

        System.out.printf("%d links discovered on %s%n", links.size(), target);
        int i = 1;
        for (Map<String, Object> link : links.subList(0, Math.min(20, links.size()))) {
            System.out.printf("[%2d] %s%n", i++, link.getOrDefault("url", ""));
        }
    }
}
