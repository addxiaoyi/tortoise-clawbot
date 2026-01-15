//! Tortoise Rust SDK - Basic Chat Example

use tortoise_sdk::{Client, ClientConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create client
    let config = ClientConfig::default()
        .with_url("http://localhost:8080")
        .with_api_key("your-api-key");
    
    let mut client = Client::new(config);
    
    // Connect
    println!("Connecting to gateway...");
    client.connect().await?;
    println!("Connected!\n");
    
    // Health check
    let health = client.health_check().await?;
    println!("Gateway status: {}", health.status);
    println!("Gateway version: {}\n", health.version);
    
    // Create session
    println!("Creating session...");
    let session = client.create_session("My First Chat").await?;
    println!("Created session: {}\n", session.id);
    
    // Send message
    println!("Sending message...");
    let response = client.chat("Hello! What can you do?").await?;
    println!("AI: {}\n", response.content);
    
    // Continue conversation
    println!("Continuing conversation...");
    let response2 = client.chat("Tell me more about yourself.").await?;
    println!("AI: {}\n", response2.content);
    
    // Get all sessions
    let sessions = client.get_sessions().await?;
    println!("Total sessions: {}", sessions.len());
    
    // List plugins
    let plugins = client.list_plugins().await?;
    println!("Available plugins: {}", plugins.len());
    
    // List channels
    let channels = client.list_channels().await?;
    println!("Available channels: {}", channels.len());
    
    println!("\nDone!");
    Ok(())
}
