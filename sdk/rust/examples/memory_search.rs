//! Tortoise Rust SDK - Memory Management Example

use tortoise_sdk::{Client, ClientConfig};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let config = ClientConfig::default()
        .with_url("http://localhost:8080")
        .with_api_key("your-api-key");
    
    let client = Client::new(config);
    
    println!("Adding memories...\n");
    
    // Add memories
    client.memory_add(
        "User prefers dark mode interface",
        "user_preference"
    ).await?;
    
    client.memory_add(
        "User is interested in AI and machine learning",
        "interest"
    ).await?;
    
    client.memory_add(
        "User works as a backend developer",
        "work"
    ).await?;
    
    println!("Added 3 memories\n");
    
    // Search memories
    println!("Searching for 'AI'...");
    let results = client.memory_search("AI", 10).await?;
    
    for entry in results {
        println!("- {} ({})", entry.content, entry.memory_type);
    }
    
    println!("\nDone!");
    Ok(())
}
