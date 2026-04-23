//! Search the web and optionally scrape each result inline.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example search -- "rust programming"

use std::env;

use zapfetch::{Client, Format, ScrapeOptions, SearchOptions, SearchResultOrDocument};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;
    let query = env::args().skip(1).collect::<Vec<_>>().join(" ");
    let query = if query.is_empty() {
        "rust programming".to_string()
    } else {
        query
    };

    let resp = client
        .search(
            &query,
            SearchOptions {
                limit: Some(5),
                scrape_options: Some(ScrapeOptions {
                    formats: Some(vec![Format::Markdown]),
                    only_main_content: Some(true),
                    ..Default::default()
                }),
                ..Default::default()
            },
        )
        .await?;

    println!("query={:?}", query);
    if let Some(web) = resp.data.web {
        for (i, item) in web.iter().enumerate() {
            match item {
                SearchResultOrDocument::WebResult(r) => {
                    println!(
                        "[{:>2}] {}\n     {}",
                        i + 1,
                        r.title.clone().unwrap_or_default(),
                        r.url
                    );
                }
                SearchResultOrDocument::Document(d) => {
                    let title = d
                        .metadata
                        .as_ref()
                        .and_then(|m| m.title.clone())
                        .unwrap_or_default();
                    println!("[{:>2}] (scraped) {}", i + 1, title);
                }
            }
        }
    }
    Ok(())
}
