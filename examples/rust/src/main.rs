// Tortoise Rust SDK - Example Usage

use tortoise::{Client, Config};

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    // Create client
    let config = Config::default()
        .with_api_key("your-api-key")
        .with_base_url("http://localhost:18792");
    
    let client = Client::new(config);

    // Connect
    client.connect().await?;

    // Create session
    let session = client
        .sessions()
        .create()
        .user_id("user@example.com")
        .model("gpt-4o")
        .temperature(0.7)
        .send()
        .await?;

    println!("Session created: {}", session.id);

    // Send message
    let response = client
        .messages()
        .send(&session.id, "Hello!")
        .await?;

    println!("Response: {}", response.content);

    // Stream response
    let mut stream = client
        .messages()
        .stream(&session.id, "Count to 3")
        .await?;

    print!("Streaming: ");
    while let Some(event) = stream.next().await {
        match event {
            tortoise::StreamEvent::ContentChunk(delta) => {
                print!("{}", delta);
            }
            tortoise::StreamEvent::MessageEnd => {
                println!();
            }
            _ => {}
        }
    }

    // List tools
    let tools = client.tools().list().await?;
    println!("Available tools: {:?}", tools);

    // Invoke tool
    let result = client
        .tools()
        .invoke("calculator")
        .arg("expression", "100 + 200")
        .send()
        .await?;

    println!("Calculator result: {}", result.result);

    // Store memory
    let memory = client
        .memory()
        .store()
        .content("User prefers dark mode")
        .memory_type(tortoise::MemoryType::Fact)
        .tags(&["ui", "theme"])
        .importance(0.8)
        .send()
        .await?;

    println!("Memory stored: {}", memory.id);

    // Search memories
    let results = client
        .memory()
        .search()
        .query("theme")
        .limit(10)
        .send()
        .await?;

    println!("Found {} memories", results.len());

    client.disconnect().await?;

    Ok(())
}
