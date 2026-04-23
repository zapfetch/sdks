# Examples

Each file here is a standalone snippet demonstrating one endpoint.

## Running

The examples live outside of the main source set so they do not ship in the
published jar. Two ways to try them:

### 1. Copy into your own project

```xml
<!-- Maven -->
<dependency>
  <groupId>com.zapfetch</groupId>
  <artifactId>zapfetch-java</artifactId>
  <version>0.1.0</version>
</dependency>
```

```kotlin
// Gradle (Kotlin DSL)
dependencies {
    implementation("com.zapfetch:zapfetch-java:0.1.0")
}
```

Then copy any `*.java` from this directory into your project's source tree,
set `ZAPFETCH_API_KEY`, and run.

### 2. Build the SDK locally and run with `java`

```bash
cd java-sdk
./gradlew jar
# produces build/libs/zapfetch-java-0.1.0.jar
# plus transitive deps resolved into ~/.gradle/caches/modules-2/...

# Shortcut: use Gradle to build a single fat jar for convenient example runs
# (not wired up here to keep the publishable jar thin)
```

The simplest path is `./gradlew :test --tests com.zapfetch.ZapfetchClientTest`
to verify the checkout builds; then copy the example you want into your own
project as in option 1.

## Files

| File | Demonstrates |
|---|---|
| `ScrapeExample.java` | Single-page scrape with format + main-content options |
| `CrawlExample.java` | Sync polling crawl with depth / limit constraints |
| `SearchExample.java` | Web search with result limit |
| `MapExample.java` | URL discovery on a domain |
| `BatchScrapeExample.java` | Concurrent scrape across many URLs |
| `AgentExample.java` | Structured extraction with a JSON schema |
