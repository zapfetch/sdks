package com.zapfetch.examples;

import com.zapfetch.client.ZapfetchClient;
import com.zapfetch.models.AgentOptions;
import com.zapfetch.models.AgentStatusResponse;

import java.util.List;
import java.util.Map;

/**
 * Run an agent task that extracts structured data with a JSON schema.
 *
 * Usage:
 *   ZAPFETCH_API_KEY=zf-... java ... com.zapfetch.examples.AgentExample
 */
public final class AgentExample {
    private AgentExample() {}

    public static void main(String[] args) {
        ZapfetchClient client = ZapfetchClient.builder()
            .apiKey(System.getenv("ZAPFETCH_API_KEY"))
            .build();

        Map<String, Object> schema = Map.of(
            "type", "object",
            "properties", Map.of(
                "plans", Map.of(
                    "type", "array",
                    "items", Map.of(
                        "type", "object",
                        "properties", Map.of(
                            "name",  Map.of("type", "string"),
                            "price", Map.of("type", "string")
                        ),
                        "required", List.of("name", "price")
                    )
                )
            ),
            "required", List.of("plans")
        );

        AgentStatusResponse status = client.agent(AgentOptions.builder()
            .urls(List.of("https://example.com/pricing"))
            .prompt("Extract each pricing plan with its price.")
            .schema(schema)
            .maxCredits(50)
            .build());

        System.out.printf("agent status=%s%n", status.getStatus());
        System.out.println(status.getData());
    }
}
