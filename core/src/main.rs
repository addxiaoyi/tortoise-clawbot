//! Tortoise CLI - 主入口
//!
//! 命令行界面

use anyhow::Result;
use clap::{Parser, Subcommand};
use tortoise_core::{Runtime, RuntimeConfig, init_logging};

#[derive(Parser)]
#[command(name = "tortoise")]
#[command(about = "Tortoise AI Agent Framework", long_about = None)]
#[command(version)]
struct Cli {
    /// 详细输出
    #[arg(short, long)]
    verbose: bool,
    
    /// 配置文件路径
    #[arg(short, long, default_value = "~/.tortoise/config.toml")]
    config: Option<String>,
    
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// 启动 Tortoise
    Start {
        /// 绑定地址
        #[arg(long, default_value = "127.0.0.1")]
        bind: String,
        
        /// 端口
        #[arg(long, default_value = "8080")]
        port: u16,
    },
    
    /// 停止 Tortoise
    Stop,
    
    /// 重启 Tortoise
    Restart,
    
    /// 状态检查
    Status {
        /// 深度检查
        #[arg(long)]
        deep: bool,
    },
    
    /// 聊天
    Chat {
        /// 消息内容
        message: String,
        
        /// 思维模式
        #[arg(long, default_value = "balanced")]
        thinking: String,
    },
    
    /// 配置管理
    Config {
        #[command(subcommand)]
        action: ConfigAction,
    },
    
    /// 插件管理
    Plugin {
        #[command(subcommand)]
        action: PluginAction,
    },
    
    /// 通道管理
    Channel {
        #[command(subcommand)]
        action: ChannelAction,
    },
}

#[derive(Subcommand)]
enum ConfigAction {
    /// 显示当前配置
    Show,
    
    /// 设置配置项
    Set {
        key: String,
        value: String,
    },
    
    /// 获取配置项
    Get {
        key: String,
    },
}

#[derive(Subcommand)]
enum PluginAction {
    /// 列出插件
    List,
    
    /// 安装插件
    Install {
        path: String,
    },
    
    /// 卸载插件
    Uninstall {
        id: String,
    },
    
    /// 启用插件
    Enable {
        id: String,
    },
    
    /// 禁用插件
    Disable {
        id: String,
    },
}

#[derive(Subcommand)]
enum ChannelAction {
    /// 列出通道
    List,
    
    /// 连接通道
    Connect {
        channel: String,
    },
    
    /// 断开通道
    Disconnect {
        channel: String,
    },
    
    /// 发送消息
    Send {
        channel: String,
        message: String,
    },
}

#[tokio::main]
async fn main() -> Result<()> {
    let cli = Cli::parse();
    
    // 初始化日志
    if cli.verbose {
        init_logging!(tracing::Level::DEBUG);
    } else {
        init_logging!();
    }
    
    tracing::info!("Tortoise AI Agent Framework v{}", tortoise_core::VERSION);
    
    match cli.command {
        Commands::Start { bind, port } => {
            tracing::info!("Starting Tortoise on {}:{}", bind, port);
            let config = RuntimeConfig::default();
            let runtime = Runtime::new(config).await?;
            runtime.start().await?;
            
            // 保持运行
            tokio::signal::ctrl_c().await?;
            
            tracing::info!("Shutting down...");
            runtime.stop().await?;
        }
        
        Commands::Stop => {
            tracing::info!("Stopping Tortoise...");
        }
        
        Commands::Restart => {
            tracing::info!("Restarting Tortoise...");
        }
        
        Commands::Status { deep } => {
            tracing::info!("Checking status...");
            if deep {
                tracing::info!("Deep check: enabled");
            }
        }
        
        Commands::Chat { message, thinking } => {
            tracing::info!("Chat mode: {}", message);
            tracing::info!("Thinking mode: {}", thinking);
        }
        
        Commands::Config { action } => {
            match action {
                ConfigAction::Show => {
                    tracing::info!("Showing configuration...");
                }
                ConfigAction::Set { key, value } => {
                    tracing::info!("Setting {} = {}", key, value);
                }
                ConfigAction::Get { key } => {
                    tracing::info!("Getting {}", key);
                }
            }
        }
        
        Commands::Plugin { action } => {
            match action {
                PluginAction::List => {
                    tracing::info!("Listing plugins...");
                }
                PluginAction::Install { path } => {
                    tracing::info!("Installing plugin from: {}", path);
                }
                PluginAction::Uninstall { id } => {
                    tracing::info!("Uninstalling plugin: {}", id);
                }
                PluginAction::Enable { id } => {
                    tracing::info!("Enabling plugin: {}", id);
                }
                PluginAction::Disable { id } => {
                    tracing::info!("Disabling plugin: {}", id);
                }
            }
        }
        
        Commands::Channel { action } => {
            match action {
                ChannelAction::List => {
                    tracing::info!("Listing channels...");
                }
                ChannelAction::Connect { channel } => {
                    tracing::info!("Connecting channel: {}", channel);
                }
                ChannelAction::Disconnect { channel } => {
                    tracing::info!("Disconnecting channel: {}", channel);
                }
                ChannelAction::Send { channel, message } => {
                    tracing::info!("Sending to {}: {}", channel, message);
                }
            }
        }
    }
    
    Ok(())
}
