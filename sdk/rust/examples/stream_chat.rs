//! Tortoise Rust SDK - Streaming Chat Example

use futures_util::StreamExt;
use tortoise_sdk::{Client, ClientConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = ClientConfig::default()
        .with_url("http://localhost:8080")
        .with_api_key("your-api-key");
    
    let client = Client::new(config);
    
    println!("Starting streaming chat...\n");
    
    // Create streaming request
    let stream = client.chat_stream("Write a haiku about Rust programming.").await?;
    
    // Process stream
    let mut stream = stream?;
    while let Some(chunk_result) = stream.next().await {
        match chunk_result {
            Ok(text) => print!("{}", text),
            Err(e) => eprintln!("Error: {}", e),
        }
    }
    
    println!("\n\n[Streaming complete]");
    Ok(())
}
