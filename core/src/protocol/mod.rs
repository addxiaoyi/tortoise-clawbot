//! Protocol implementation for Tortoise
//!
//! Implements the Tortoise binary protocol for high-performance communication.

use serde::{Deserialize, Serialize};
use uuid::Uuid;
use chrono::{DateTime, Utc};

pub mod codec;
pub mod messages;

pub use codec::ProtocolCodec;
pub use messages::*;

/// Message types in Tortoise Protocol
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
#[repr(u16)]
pub enum MessageType {
    Handshake = 0x0001,
    HandshakeAck = 0x0002,
    Request = 0x0003,
    Response = 0x0004,
    StreamStart = 0x0005,
    StreamChunk = 0x0006,
    StreamEnd = 0x0007,
    Event = 0x0008,
    ToolCall = 0x0009,
    ToolResult = 0x000A,
    Error = 0x000B,
    Heartbeat = 0x000C,
    Close = 0x000D,
}

impl TryFrom<u16> for MessageType {
    type Error = super::Error