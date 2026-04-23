//! Discover URLs on a site using the map endpoint.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example map -- https://example.com pricing

use std::env;

use zapfetch::{Client, MapOptions, SitemapMode};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;
    let mut args = env::args().skip(1);
    let target = args.next().unwrap_or_else(|| "https://example.com".into());
    let search = args.next();

    let resp = client
        .map(
            &target,
            MapOptions {
                limit: Some(50),
                sitemap: Some(SitemapMode::Include),
                search,
                ..Default::default()
            },
        )
        .await?;

    println!("{} links discovered on {}", resp.links.len(), target);
    for (i, link) in resp.links.iter().take(20).enumerate() {
        println!("[{:>2}] {}", i + 1, link.url);
    }
    Ok(())
}
