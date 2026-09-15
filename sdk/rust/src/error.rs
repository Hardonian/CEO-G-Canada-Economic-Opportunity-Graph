use thiserror::Error;

/// All errors that can be returned by the COG SDK.
#[derive(Debug, Error)]
pub enum CogError {
    /// An HTTP transport-level error from `reqwest`.
    #[error("HTTP error: {0}")]
    Http(#[from] reqwest::Error),

    /// The API returned a non-2xx status code with a structured error body.
    #[error("API error {status}: {message}")]
    Api {
        /// HTTP status code.
        status: u16,
        /// Machine-readable error code from the `error` field.
        code: String,
        /// Human-readable error description.
        message: String,
    },

    /// The response body could not be deserialized.
    #[error("Decode error: {0}")]
    Decode(#[from] serde_json::Error),

    /// The base URL provided to the client is invalid.
    #[error("Invalid base URL: {0}")]
    InvalidBaseUrl(String),
}

/// A convenience `Result` alias for COG SDK operations.
pub type CogResult<T> = std::result::Result<T, CogError>;
