//! Scrape a single URL and print its markdown.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example scrape -- https://example.com

use std::env;

use zapfetch::{Client, Format, ScrapeOptions};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;
    let target = env::args().nth(1).unwrap_or_else(|| "https://example.com".into());

    let doc = client
        .scrape(
            &target,
            ScrapeOptions {
                formats: Some(vec![Format::Markdown]),
                only_main_content: Some(true),
                ..Default::default()
            },
        )
        .await?;

    println!("--- {} ---", target);
    if let Some(md) = doc.markdown {
        println!("{}", md);
    }
    Ok(())
}
