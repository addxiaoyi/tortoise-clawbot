//! Tortoise Core Runtime - 核心运行时库
//! 
//! 高性能、可扩展的 AI Agent 运行时核心

pub mod runtime;
pub mod protocol;
pub mod memory;
pub mod plugin;
pub mod session;
pub mod tool;
pub mod channel;
pub mod error;

pub use runtime::Runtime;
pub use error::{Error, Result};
