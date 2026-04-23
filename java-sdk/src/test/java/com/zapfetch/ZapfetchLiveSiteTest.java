package com.zapfetch;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.Document;
import com.zapfetch.models.ScrapeOptions;
import org.junit.jupiter.api.BeforeAll;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.condition.EnabledIfEnvironmentVariable;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

/**
 * Live Site Test - Zapfetch.dev
 * 
 * Tests the Java SDK against the actual Zapfetch production website.
 * This demonstrates real-world usage of the API against live content.
 * 
 * Run with: ZAPFETCH_API_KEY=fc-xxx gradle test --tests "com.zapfetch.ZapfetchLiveSiteTest"
 */
class ZapfetchLiveSiteTest {

    private static ZapfetchClient client;

    @BeforeAll
    static void setup() {
        String apiKey = System.getenv("ZAPFETCH_API_KEY");
        if (apiKey != null && !apiKey.isBlank()) {
            client = ZapfetchClient.fromEnv();
        }
    }

    @Test
    @EnabledIfEnvironmentVariable(named = "ZAPFETCH_API_KEY", matches = ".*\\S.*")
    void testScrapeZapfetchHomepage() {
        System.out.println("\n=== Testing against LIVE Zapfetch.dev website ===\n");
        System.out.println("Scraping: https://zapfetch.com");
        
        Document doc = client.scrape("https://zapfetch.com",
                ScrapeOptions.builder()
                        .formats(List.of("markdown", "html"))
                        .onlyMainContent(true)
                        .build());

        // Assertions
        assertNotNull(doc, "Document should not be null");
        assertNotNull(doc.getMarkdown(), "Markdown content should not be null");
        assertNotNull(doc.getHtml(), "HTML content should not be null");
        assertNotNull(doc.getMetadata(), "Metadata should not be null");
        
        // Verify it's actually the Zapfetch site
        String markdown = doc.getMarkdown().toLowerCase();
        assertTrue(markdown.contains("zapfetch") || markdown.contains("scrape") || markdown.contains("crawl"),
                "Content should mention Zapfetch features");
        
        // Check metadata
        String sourceUrl = doc.getMetadata().get("sourceURL").toString();
        assertTrue(sourceUrl.contains("zapfetch.com"), "Source URL should be zapfetch.com");
        
        // Display results
        System.out.println("\n✓ Successfully scraped Zapfetch.dev!");
        System.out.println("\nMetadata:");
        System.out.println("  Source URL: " + sourceUrl);
        if (doc.getMetadata().get("title") != null) {
            System.out.println("  Title: " + doc.getMetadata().get("title"));
        }
        System.out.println("  Status Code: " + doc.getMetadata().get("statusCode"));
        
        System.out.println("\nContent Stats:");
        System.out.println("  Markdown length: " + doc.getMarkdown().length() + " characters");
        System.out.println("  HTML length: " + doc.getHtml().length() + " characters");
        
        System.out.println("\nFirst 500 characters of markdown:");
        System.out.println("  " + doc.getMarkdown().substring(0, Math.min(500, doc.getMarkdown().length())).replace("\n", "\n  "));
        
        System.out.println("\n=== Live site test completed successfully! ===\n");
    }

    @Test
    @EnabledIfEnvironmentVariable(named = "ZAPFETCH_API_KEY", matches = ".*\\S.*")
    void testScrapeZapfetchPricing() {
        System.out.println("\n=== Testing Zapfetch Pricing Page ===\n");
        System.out.println("Scraping: https://zapfetch.com/pricing");
        
        Document doc = client.scrape("https://zapfetch.com/pricing",
                ScrapeOptions.builder()
                        .formats(List.of("markdown"))
                        .build());

        // Assertions
        assertNotNull(doc, "Document should not be null");
        assertNotNull(doc.getMarkdown(), "Markdown content should not be null");
        
        String markdown = doc.getMarkdown().toLowerCase();
        assertTrue(markdown.contains("pricing") || markdown.contains("plan") || markdown.contains("price"),
                "Pricing page should contain pricing information");
        
        System.out.println("✓ Successfully scraped pricing page!");
        System.out.println("  Content length: " + doc.getMarkdown().length() + " characters");
        System.out.println("  Source: " + doc.getMetadata().get("sourceURL"));
    }

    @Test
    @EnabledIfEnvironmentVariable(named = "ZAPFETCH_API_KEY", matches = ".*\\S.*")
    void testScrapeZapfetchDocs() {
        System.out.println("\n=== Testing Zapfetch Documentation ===\n");
        System.out.println("Scraping: https://docs.zapfetch.com");
        
        Document doc = client.scrape("https://docs.zapfetch.com",
                ScrapeOptions.builder()
                        .formats(List.of("markdown"))
                        .waitFor(2000)  // Wait for docs to load
                        .build());

        // Assertions
        assertNotNull(doc, "Document should not be null");
        assertNotNull(doc.getMarkdown(), "Markdown content should not be null");
        assertFalse(doc.getMarkdown().isEmpty(), "Markdown should not be empty");
        
        String markdown = doc.getMarkdown().toLowerCase();
        assertTrue(markdown.contains("document") || markdown.contains("api") || markdown.contains("zapfetch"),
                "Docs should contain documentation content");
        
        System.out.println("✓ Successfully scraped documentation!");
        System.out.println("  Content length: " + doc.getMarkdown().length() + " characters");
        System.out.println("  Source: " + doc.getMetadata().get("sourceURL"));
        
        System.out.println("\n=== All Zapfetch.dev tests passed! ===\n");
    }
}
