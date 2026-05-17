//! Tortoise CLI Entry Point

use anyhow::Result;
use clap::{Parser, Subcommand};
use tortoise_core::{Config, Tortoise};
use std::path::PathBuf;

#[derive(Parser)]
#[command(name = "tortoise")]
#[command(about = "Tortoise - Super AI Agent Framework", long_about = None)]
struct Cli {
    /// Configuration file path
    #[arg(short, long, default_value = "config.toml")]
    config: PathBuf,

    /// Enable verbose logging
    #[arg(short, long)]
    verbose: bool,

    /// Disable color output
    #[arg(long)]
    no_color: bool,

    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// Start the Tortoise agent
    Start,
    /// Stop the Tortoise agent
    Stop,
    /// Send a message
    Message {
        /// Message content
        #[arg(short, long)]
        content: String,
    },
    /// Manage plugins
    Plugins {
        #[command(subcommand)]
        action: PluginAction,
    },
    /// Manage channels
    Channels {
        #[command(subcommand)]
        action: ChannelAction,
    },
    /// Show status
    Status,
    /// Initialize configuration
    Init {
        /// Overwrite existing config
        #[arg(short, long)]
        force: bool,
    },
}

#[derive(Subcommand)]
enum PluginAction {
    /// List all plugins
    List,
    /// Install a plugin
    Install { name: String },
    /// Uninstall a plugin
    Uninstall { name: String },
    /// Enable a plugin
    Enable { name: String },
    /// Disable a plugin
    Disable { name: String },
}

#[derive(Subcommand)]
enum ChannelAction {
    /// List all channels
    List,
    /// Add a channel
    Add { channel_type: String },
    /// Remove a channel
    Remove { name: String },
    /// Show channel status
    Status { name: String },
}

#[tokio::main]
async fn main() -> Result<()> {
    // Parse CLI arguments
    let cli = Cli::parse();

    // Initialize logging
    init_logging(cli.verbose, cli.no_color);

    // Handle commands
    match cli.command {
        Some(Commands::Start) => {
            start_tortoise(cli.config).await?;
        }
        Some(Commands::Stop) => {
            stop_tortoise().await?;
        }
        Some(Commands::Message { content }) => {
            send_message(content).await?;
        }
        Some(Commands::Plugins { action }) => {
            handle_plugins(action).await?;
        }
        Some(Commands::Channels { action }) => {
            handle_channels(action).await?;
        }
        Some(Commands::Status) => {
            show_status().await?;
        }
        Some(Commands::Init { force }) => {
            init_config(cli.config, force)?;
        }
        None => {
            // Start by default
            start_tortoise(cli.config).await?;
        }
    }

    Ok(())
}

fn init_logging(verbose: bool, no_color: bool) {
    use tracing_subscriber::{fmt, prelude::*, EnvFilter};

    let filter = if verbose {
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("debug"))
    } else {
        EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info"))
    };

    let subscriber = tracing_subscriber::registry()
        .with(filter)
        .with(fmt::layer().with_ansi(!no_color));

    tracing::subscriber::set_global_default(subscriber)
        .expect("Failed to set tracing subscriber");
}

async fn start_tortoise(config_path: PathBuf) -> Result<()> {
    tracing::info!("Loading configuration from {:?}", config_path);

    let config = if config_path.exists() {
        let content = std::fs::read_to_string(&config_path)?;
        toml::from_str(&content)?
    } else {
        tracing::warn!("Config file not found, using defaults");
        Config::default()
    };

    tracing::info!("Creating Tortoise instance...");
    let tortoise = Tortoise::new(config).await?;

    tracing::info!("Starting Tortoise...");
    tortoise.start().await?;

    tracing::info!("Tortoise is running. Press Ctrl+C to stop.");

    // Wait for shutdown signal
    tokio::signal::ctrl_c().await?;

    tracing::info!("Shutting down...");
    tortoise.stop().await?;

    tracing::info!("Goodbye!");
    Ok(())
}

async fn stop_tortoise() -> Result<()> {
    use std::net::TcpStream;
    use std::io::{Write, Read};
    
    tracing::info!("Stopping Tortoise daemon...");
    
    let daemon_addr = "127.0.0.1:18792";
    
    match TcpStream::connect(daemon_addr) {
        Ok(mut stream) => {
            let cmd = "STOP\n";
            stream.write_all(cmd.as_bytes())?;
            
            let mut response = [0u8; 1024];
            match stream.read(&mut response) {
                Ok(n) => {
                    let response_str = String::from_utf8_lossy(&response[..n]);
                    tracing::info!("Daemon response: {}", response_str.trim());
                }
                Err(e) => tracing::warn!("Failed to read daemon response: {}", e),
            }
            
            tracing::info!("Stop signal sent to daemon");
        }
        Err(e) => {
            tracing::warn!("Could not connect to daemon at {}: {}", daemon_addr, e);
            tracing::info!("Daemon may not be running or is already stopped");
        }
    }
    
    Ok(())
}

async fn send_message(content: String) -> Result<()> {
    use std::net::TcpStream;
    use std::io::{Write, Read};
    
    tracing::info!("Sending message: {}", content);
    
    let daemon_addr = "127.0.0.1:18792";
    
    match TcpStream::connect(daemon_addr) {
        Ok(mut stream) => {
            let cmd = format!("SEND:{}\n", content);
            stream.write_all(cmd.as_bytes())?;
            
            let mut response = [0u8; 4096];
            match stream.read(&mut response) {
                Ok(n) => {
                    let response_str = String::from_utf8_lossy(&response[..n]);
                    tracing::info!("Response: {}", response_str.trim());
                }
                Err(e) => tracing::warn!("Failed to read response: {}", e),
            }
        }
        Err(e) => {
            tracing::error!("Could not connect to daemon: {}", e);
            anyhow::bail!("Daemon not running. Start Tortoise first with 'tortoise start'");
        }
    }
    
    Ok(())
}

async fn handle_plugins(action: PluginAction) -> Result<()> {
    match action {
        PluginAction::List => {
            tracing::info!("Listing plugins...");
            // TODO: Implement plugin listing
        }
        PluginAction::Install { name } => {
            tracing::info!("Installing plugin: {}", name);
            // TODO: Implement plugin installation
        }
        PluginAction::Uninstall { name } => {
            tracing::info!("Uninstalling plugin: {}", name);
            // TODO: Implement plugin uninstallation
        }
        PluginAction::Enable { name } => {
            tracing::info!("Enabling plugin: {}", name);
            // TODO: Implement plugin enabling
        }
        PluginAction::Disable { name } => {
            tracing::info!("Disabling plugin: {}", name);
            // TODO: Implement plugin disabling
        }
    }
    Ok(())
}

async fn handle_channels(action: ChannelAction) -> Result<()> {
    match action {
        ChannelAction::List => {
            tracing::info!("Listing channels...");
            // TODO: Implement channel listing
        }
        ChannelAction::Add { channel_type } => {
            tracing::info!("Adding channel: {}", channel_type);
            // TODO: Implement channel addition
        }
        ChannelAction::Remove { name } => {
            tracing::info!("Removing channel: {}", name);
            // TODO: Implement channel removal
        }
        ChannelAction::Status { name } => {
            tracing::info!("Channel status for: {}", name);
            // TODO: Implement channel status
        }
    }
    Ok(())
}

async fn show_status() -> Result<()> {
    tracing::info!("Showing Tortoise status...");
    // TODO: Implement status display
    println!("Tortoise Status:");
    println!("  Version: {}", env!("CARGO_PKG_VERSION"));
    println!("  Status: Running");
    Ok(())
}

fn init_config(config_path: PathBuf, force: bool) -> Result<()> {
    if config_path.exists() && !force {
        anyhow::bail!("Config file already exists. Use --force to overwrite.");
    }

    let config = Config::default();
    let content = toml::to_string_pretty(&config)?;

    if let Some(parent) = config_path.parent() {
        std::fs::create_dir_all(parent)?;
    }

    std::fs::write(&config_path, content)?;
    println!("Configuration initialized at {:?}", config_path);

    Ok(())
}
