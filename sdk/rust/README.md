# Tortoise Rust SDK

Rust SDK for Tortoise AI Agent Framework.

## Installation

Add to your `Cargo.toml`:

```toml
[dependencies]
tortoise-sdk = "0.1"
```

## Quick Start

```rust
use tortoise_sdk::{Client, Config};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let client = Client::new(Config {
        api_key: "your-api-key".to_string(),
        base_url: "http://localhost:8080".to_string(),
    })?;

    // Create session
    let session = client.create_session("Hello Session").await?;
    println!("Session: {:?}", session.id);

    // Send message
    let response = session.chat("Hello, Tortoise!").await?;
    println!("Response: {}", response.content);

    Ok(())
}
```

## Features

- Session management
- Multi-agent orchestration
- Plugin marketplace
- Memory system
- Channel integration
- Enterprise authentication

## Documentation

See [docs/](docs/) for full documentation.
