//! Batch-scrape a list of URLs concurrently and print basic metadata.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example batch_scrape

use zapfetch::{BatchScrapeOptions, Client, Format, ScrapeOptions};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;

    let urls = vec![
        "https://example.com".to_string(),
        "https://example.org".to_string(),
        "https://example.net".to_string(),
    ];

    let job = client
        .batch_scrape(
            urls,
            BatchScrapeOptions {
                options: Some(ScrapeOptions {
                    formats: Some(vec![Format::Markdown]),
                    only_main_content: Some(true),
                    ..Default::default()
                }),
                max_concurrency: Some(3),
                poll_interval: Some(2000),
                ..Default::default()
            },
        )
        .await?;

    println!(
        "batch status={:?} completed={}/{} credits={:?}",
        job.status, job.completed, job.total, job.credits_used
    );
    for (i, doc) in job.data.iter().enumerate() {
        let title = doc
            .metadata
            .as_ref()
            .and_then(|m| m.title.clone())
            .unwrap_or_default();
        let source = doc
            .metadata
            .as_ref()
            .and_then(|m| m.source_url.clone())
            .unwrap_or_default();
        println!("[{:>2}] {}\n     {}", i + 1, title, source);
    }
    Ok(())
}
