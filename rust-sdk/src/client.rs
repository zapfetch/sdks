//! ZapFetch API v2 client.

use reqwest::Response;
use serde::de::DeserializeOwned;
use serde_json::Value;

use crate::error::{ApiErrorBody, Error};

pub(crate) const API_VERSION: &str = "/v2";
const CLOUD_API_URL: &str = "https://api.zapfetch.com";

/// ZapFetch API v2 client.
///
/// This client provides access to all v2 API endpoints including scrape, crawl,
/// search, map, batch scrape, and agent operations.
///
/// # Example
///
/// ```no_run
/// use zapfetch::Client;
///
/// #[tokio::main]
/// async fn main() -> Result<(), Box<dyn std::error::Error>> {
///     // Create a client for the ZapFetch cloud service
///     let client = Client::new("your-api-key")?;
///
///     // Or create a client for a self-hosted instance
///     let client = Client::new_selfhosted("http://localhost:3000", Some("api-key"))?;
///
///     Ok(())
/// }
/// ```
#[derive(Clone, Debug)]
pub struct Client {
    pub(crate) api_key: Option<String>,
    pub(crate) api_url: String,
    pub(crate) client: reqwest::Client,
}

impl Client {
    /// Creates a new client for the ZapFetch cloud service.
    ///
    /// # Arguments
    ///
    /// * `api_key` - Your ZapFetch API key.
    ///
    /// # Errors
    ///
    /// Returns an error if the API key is empty.
    ///
    /// # Example
    ///
    /// ```no_run
    /// use zapfetch::Client;
    ///
    /// let client = Client::new("your-api-key").unwrap();
    /// ```
    pub fn new(api_key: impl AsRef<str>) -> Result<Self, Error> {
        Client::new_selfhosted(CLOUD_API_URL, Some(api_key))
    }

    /// Creates a new client for a self-hosted ZapFetch instance.
    ///
    /// # Arguments
    ///
    /// * `api_url` - The base URL of your ZapFetch instance.
    /// * `api_key` - Optional API key (required for cloud, optional for self-hosted).
    ///
    /// # Errors
    ///
    /// Returns an error if using the cloud service without an API key.
    ///
    /// # Example
    ///
    /// ```no_run
    /// use zapfetch::Client;
    ///
    /// // Self-hosted without authentication
    /// let client = Client::new_selfhosted("http://localhost:3000", None::<&str>).unwrap();
    ///
    /// // Self-hosted with authentication
    /// let client = Client::new_selfhosted("http://localhost:3000", Some("api-key")).unwrap();
    /// ```
    pub fn new_selfhosted(
        api_url: impl AsRef<str>,
        api_key: Option<impl AsRef<str>>,
    ) -> Result<Self, Error> {
        // Normalize URL by trimming trailing slashes for consistent comparison
        let url = api_url.as_ref().trim_end_matches('/').to_string();
        let api_key = api_key.map(|k| k.as_ref().to_string());

        // Reject empty or missing API key for cloud service
        if url == CLOUD_API_URL {
            match &api_key {
                None => {
                    return Err(Error::Api(
                        "Configuration".to_string(),
                        ApiErrorBody {
                            success: false,
                            error: "API key is required for cloud service".to_string(),
                            details: None,
                        },
                    ));
                }
                Some(key) if key.trim().is_empty() => {
                    return Err(Error::Api(
                        "Configuration".to_string(),
                        ApiErrorBody {
                            success: false,
                            error: "API key cannot be empty for cloud service".to_string(),
                            details: None,
                        },
                    ));
                }
                _ => {}
            }
        }

        Ok(Client {
            api_key,
            api_url: url,
            client: reqwest::Client::new(),
        })
    }

    /// Creates a client from environment variables.
    ///
    /// Reads `ZAPFETCH_API_KEY` (required) and `ZAPFETCH_API_URL` (optional;
    /// defaults to `https://api.zapfetch.com`).
    ///
    /// # Errors
    ///
    /// Returns an error if `ZAPFETCH_API_KEY` is unset or empty.
    pub fn from_env() -> Result<Self, Error> {
        let api_key = std::env::var("ZAPFETCH_API_KEY").unwrap_or_default();
        if api_key.trim().is_empty() {
            return Err(Error::Api(
                "Configuration".to_string(),
                ApiErrorBody {
                    success: false,
                    error: "ZAPFETCH_API_KEY environment variable is required".to_string(),
                    details: None,
                },
            ));
        }
        let api_url =
            std::env::var("ZAPFETCH_API_URL").unwrap_or_else(|_| CLOUD_API_URL.to_string());
        Client::new_selfhosted(api_url, Some(api_key))
    }

    /// Creates a client with a user-supplied `reqwest::Client`.
    ///
    /// Use this to customise timeouts, proxy behaviour, TLS roots, or any other
    /// [`reqwest::ClientBuilder`] option. For example, to bypass a system proxy
    /// during local development:
    ///
    /// ```no_run
    /// use zapfetch::Client;
    ///
    /// let http = reqwest::Client::builder().no_proxy().build().unwrap();
    /// let client = Client::with_http_client("https://api.zapfetch.com", Some("zf-key"), http).unwrap();
    /// ```
    pub fn with_http_client(
        api_url: impl AsRef<str>,
        api_key: Option<impl AsRef<str>>,
        http_client: reqwest::Client,
    ) -> Result<Self, Error> {
        let mut c = Client::new_selfhosted(api_url, api_key)?;
        c.client = http_client;
        Ok(c)
    }

    /// Prepares headers for API requests.
    pub(crate) fn prepare_headers(
        &self,
        idempotency_key: Option<&String>,
    ) -> reqwest::header::HeaderMap {
        use reqwest::header::HeaderValue;

        let mut headers = reqwest::header::HeaderMap::new();
        // Static string is always valid ASCII
        headers.insert("Content-Type", HeaderValue::from_static("application/json"));
        if let Some(api_key) = self.api_key.as_ref() {
            // API key is validated at client creation, so this should always succeed.
            // Use if-let to gracefully handle edge cases without panicking.
            if let Ok(value) = format!("Bearer {}", api_key).parse() {
                headers.insert("Authorization", value);
            }
        }
        if let Some(key) = idempotency_key {
            // Gracefully skip invalid idempotency keys instead of panicking
            if let Ok(value) = key.parse() {
                headers.insert("x-idempotency-key", value);
            }
        }
        headers
    }

    /// Handles API responses, parsing JSON and handling errors.
    pub(crate) async fn handle_response<T: DeserializeOwned>(
        &self,
        response: Response,
        action: impl AsRef<str>,
    ) -> Result<T, Error> {
        let (is_success, status) = (response.status().is_success(), response.status());

        let response = response
            .text()
            .await
            .map_err(Error::ResponseParseErrorText)
            .and_then(|response_json| {
                serde_json::from_str::<Value>(&response_json).map_err(Error::ResponseParseError)
            })
            .and_then(|response_value| {
                // Check for success field, or allow responses without it for status checks
                if action.as_ref().contains("status")
                    || action.as_ref().contains("cancel")
                    || response_value["success"].as_bool().unwrap_or(false)
                    || response_value.get("success").is_none()
                {
                    serde_json::from_value::<T>(response_value).map_err(Error::ResponseParseError)
                } else {
                    Err(Error::Api(
                        action.as_ref().to_string(),
                        serde_json::from_value(response_value)
                            .map_err(Error::ResponseParseError)?,
                    ))
                }
            });

        match &response {
            Ok(_) => response,
            Err(Error::ResponseParseError(_)) | Err(Error::ResponseParseErrorText(_)) => {
                if is_success {
                    response
                } else {
                    Err(Error::HttpRequestFailed(
                        action.as_ref().to_string(),
                        status.as_u16(),
                        status.as_str().to_string(),
                    ))
                }
            }
            Err(_) => response,
        }
    }

    /// Builds the full URL for an API endpoint.
    pub(crate) fn url(&self, path: &str) -> String {
        format!("{}{}{}", self.api_url, API_VERSION, path)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_new_client() {
        let client = Client::new("test-api-key").unwrap();
        assert_eq!(client.api_key, Some("test-api-key".to_string()));
        assert_eq!(client.api_url, CLOUD_API_URL);
    }

    #[test]
    fn test_new_client_requires_api_key_for_cloud() {
        let result = Client::new_selfhosted(CLOUD_API_URL, None::<&str>);
        assert!(result.is_err());
    }

    #[test]
    fn test_new_client_rejects_empty_api_key_for_cloud() {
        let result = Client::new_selfhosted(CLOUD_API_URL, Some(""));
        assert!(result.is_err());

        let result = Client::new_selfhosted(CLOUD_API_URL, Some("   "));
        assert!(result.is_err());
    }

    #[test]
    fn test_new_selfhosted_client() {
        let client = Client::new_selfhosted("http://localhost:3000", Some("api-key")).unwrap();
        assert_eq!(client.api_key, Some("api-key".to_string()));
        assert_eq!(client.api_url, "http://localhost:3000");
    }

    #[test]
    fn test_selfhosted_without_api_key() {
        let client = Client::new_selfhosted("http://localhost:3000", None::<&str>).unwrap();
        assert_eq!(client.api_key, None);
        assert_eq!(client.api_url, "http://localhost:3000");
    }

    #[test]
    fn test_url_builder() {
        let client = Client::new("test-key").unwrap();
        assert_eq!(client.url("/scrape"), "https://api.zapfetch.com/v2/scrape");
    }

    #[test]
    fn test_url_normalization_trailing_slash() {
        // Cloud URL with trailing slash should still require API key
        let result = Client::new_selfhosted("https://api.zapfetch.com/", None::<&str>);
        assert!(result.is_err());

        // Should work with API key
        let client = Client::new_selfhosted("https://api.zapfetch.com/", Some("key")).unwrap();
        assert_eq!(client.api_url, "https://api.zapfetch.com");

        // Self-hosted URL normalization
        let client = Client::new_selfhosted("http://localhost:3000/", None::<&str>).unwrap();
        assert_eq!(client.api_url, "http://localhost:3000");
    }

    #[test]
    fn test_from_env_missing_key() {
        // SAFETY: single-test access; std::env APIs are unsafe in edition 2024.
        // We guard the happy path separately to avoid cross-test env races.
        // Here we only assert the error when the key is unset.
        let prev = std::env::var("ZAPFETCH_API_KEY").ok();
        std::env::remove_var("ZAPFETCH_API_KEY");
        let result = Client::from_env();
        assert!(result.is_err());
        if let Some(v) = prev {
            std::env::set_var("ZAPFETCH_API_KEY", v);
        }
    }

    #[test]
    fn test_with_http_client_overrides_reqwest() {
        let http = reqwest::Client::builder().no_proxy().build().unwrap();
        let client =
            Client::with_http_client("http://localhost:3000", Some("api-key"), http).unwrap();
        assert_eq!(client.api_url, "http://localhost:3000");
        assert_eq!(client.api_key, Some("api-key".to_string()));
    }
}
