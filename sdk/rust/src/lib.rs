//! Tortoise Rust SDK
//! 
//! A Rust client library for the Tortoise AI Agent Framework.
//!
//! # Example
//!
//! ```rust,no_run
//! use tortoise_sdk::{Client, ClientConfig};
//!
//! #[tokio::main]
//! async fn main() -> Result<(), Box<dyn std::error::Error>> {
//!     let config = ClientConfig::default()
//!         .with_url("http://localhost:8080")
//!         .with_api_key("your-api-key");
//!     
//!     let client = Client::new(config);
//!     client.connect().await?;
//!     
//!     let response = client.chat("Hello!").await?;
//!     println!("{}", response.content);
//!     
//!     Ok(())
//! }
//! ```

pub mod client;
pub mod models;
pub mod error;

pub use client::Client;
pub use client::ClientConfig;
pub use models::*;
pub use error::{Result, TortoiseError};
