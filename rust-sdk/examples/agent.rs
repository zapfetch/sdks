//! Run an agent task with a typed extraction schema.
//!
//! Run with:
//!   ZAPFETCH_API_KEY=zf-... cargo run --example agent

use serde::Deserialize;
use serde_json::json;
use zapfetch::Client;

#[derive(Debug, Deserialize)]
#[allow(dead_code)]
struct CompanyInfo {
    name: String,
    description: Option<String>,
    industry: Option<String>,
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::from_env()?;

    let schema = json!({
        "type": "object",
        "properties": {
            "name":        { "type": "string" },
            "description": { "type": "string" },
            "industry":    { "type": "string" }
        },
        "required": ["name"]
    });

    let result: Option<CompanyInfo> = client
        .agent_with_schema(
            vec!["https://example.com".to_string()],
            "Extract company information from this website",
            schema,
        )
        .await?;

    match result {
        Some(info) => {
            println!("name        = {}", info.name);
            println!("description = {:?}", info.description);
            println!("industry    = {:?}", info.industry);
        }
        None => println!("no data extracted"),
    }
    Ok(())
}
