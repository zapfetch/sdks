//! Crawl a domain with depth + page-count limits, then summarise each result.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example crawl -- https://example.com

use std::env;

use zapfetch::{Client, CrawlOptions, Format, ScrapeOptions};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;
    let target = env::args().nth(1).unwrap_or_else(|| "https://example.com".into());

    let job = client
        .crawl(
            &target,
            CrawlOptions {
                limit: Some(10),
                max_discovery_depth: Some(2),
                scrape_options: Some(ScrapeOptions {
                    formats: Some(vec![Format::Markdown]),
                    only_main_content: Some(true),
                    ..Default::default()
                }),
                poll_interval: Some(3000),
                ..Default::default()
            },
        )
        .await?;

    println!(
        "crawl status={:?} completed={} total={} credits={:?}",
        job.status, job.completed, job.total, job.credits_used
    );
    for (i, doc) in job.data.iter().enumerate() {
        let title = doc
            .metadata
            .as_ref()
            .and_then(|m| m.title.clone())
            .unwrap_or_default();
        let len = doc.markdown.as_deref().map_or(0, str::len);
        println!("[{:>2}] {:>6} chars  {}", i + 1, len, title);
    }
    Ok(())
}
