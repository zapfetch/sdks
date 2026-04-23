/**
 * Zapfetch Java SDK — a type-safe client for the Zapfetch v2 web scraping API.
 *
 * <p>Quick start:
 * <pre>{@code
 * import com.zapfetch.client.ZapfetchClient;
 * import com.zapfetch.models.*;
 *
 * ZapfetchClient client = ZapfetchClient.builder()
 *     .apiKey("fc-your-api-key")
 *     .build();
 *
 * Document doc = client.scrape("https://example.com",
 *     ScrapeOptions.builder()
 *         .formats(List.of("markdown"))
 *         .build());
 *
 * System.out.println(doc.getMarkdown());
 * }</pre>
 *
 * @see com.zapfetch.client.ZapfetchClient
 */
package com.zapfetch;
