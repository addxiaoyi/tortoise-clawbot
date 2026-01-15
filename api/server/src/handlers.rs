//! API Handlers

use crate::AppState;
use anyhow::Result;
use axum::{
    extract::{Query, State, WebSocket},
    http::StatusCode,
    response::Json,
};
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_http::trace::TraceLayer;

#[derive(Debug, Serialize)]
pub struct HealthResponse {
    pub status: String,
    pub version: String,
}

pub async fn health() -> Json<HealthResponse> {
    Json(HealthResponse {
        status: "ok".to_string(),
        version: env!("CARGO_PKG_VERSION").to_string(),
    })
}

#[derive(Debug, Deserialize)]
pub struct ChatRequest {
    pub messages: Vec<MessageInput>,
    #[serde(default)]
    pub thinking_mode: Option<String>,
    #[serde(default)]
    pub temperature: Option<f32>,
    #[serde(default)]
    pub max_tokens: Option<usize>,
    #[serde(default)]
    pub stream: Option<bool>,
}

#[derive(Debug, Deserialize)]
pub struct MessageInput {
    pub role: String,
    pub content: String,
}

#[derive(Debug, Serialize)]
pub struct ChatResponse {
    pub content: String,
    pub usage: Option<TokenUsage>,
}

#[derive(Debug, Serialize)]
pub struct TokenUsage {
    pub prompt_tokens: u32,
    pub completion_tokens: u32,
    pub total_tokens: u32,
}

pub async fn chat(
    State(state): State<AppState>,
    Json(request): Json<ChatRequest>,
) -> Result<Json<ChatResponse>, StatusCode> {
    // TODO: Implement chat logic
    Ok(Json(ChatResponse {
        content: "Tortoise API chat response".to_string(),
        usage: Some(TokenUsage {
            prompt_tokens: 10,
            completion_tokens: 20,
            total_tokens: 30,
        }),
    }))
}

#[derive(Debug, Deserialize)]
pub struct MemoryQueryParams {
    pub query: Option<String>,
    #[serde(default)]
    pub limit: Option<usize>,
}

#[derive(Debug, Serialize)]
pub struct MemoryResponse {
    pub memories: Vec<MemoryItem>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct MemoryItem {
    pub id: String,
    pub content: String,
    pub importance: f32,
    pub memory_type: String,
    pub created_at: String,
}

pub async fn get_memory(
    State(_state): State<AppState>,
    Query(params): Query<MemoryQueryParams>,
) -> Result<Json<MemoryResponse>, StatusCode> {
    // TODO: Implement memory retrieval
    Ok(Json(MemoryResponse {
        memories: vec![],
    }))
}

#[derive(Debug, Deserialize)]
pub struct StoreMemoryRequest {
    pub content: String,
    pub importance: f32,
}

#[derive(Debug, Serialize)]
pub struct StoreMemoryResponse {
    pub id: String,
    pub success: bool,
}

pub async fn store_memory(
    State(_state): State<AppState>,
    Json(request): Json<StoreMemoryRequest>,
) -> Result<Json<StoreMemoryResponse>, StatusCode> {
    // TODO: Implement memory storage
    Ok(Json(StoreMemoryResponse {
        id: format!("mem_{}", uuid::Uuid::new_v4()),
        success: true,
    }))
}

#[derive(Debug, Serialize)]
pub struct MemoryStatsResponse {
    pub short_term: usize,
    pub medium_term: usize,
    pub long_term: usize,
    pub total: usize,
}

pub async fn memory_stats(
    State(_state): State<AppState>,
) -> Result<Json<MemoryStatsResponse>, StatusCode> {
    // TODO: Implement memory stats
    Ok(Json(MemoryStatsResponse {
        short_term: 5,
        medium_term: 50,
        long_term: 200,
        total: 255,
    }))
}

#[derive(Debug, Serialize)]
pub struct StatusResponse {
    pub state: String,
    pub model: String,
    pub uptime_seconds: u64,
    pub memory_stats: MemoryStatsResponse,
}

pub async fn status(
    State(_state): State<AppState>,
) -> Result<Json<StatusResponse>, StatusCode> {
    Ok(Json(StatusResponse {
        state: "running".to_string(),
        model: "gpt-4".to_string(),
        uptime_seconds: 3600,
        memory_stats: MemoryStatsResponse {
            short_term: 5,
            medium_term: 50,
            long_term: 200,
            total: 255,
        },
    }))
}

#[derive(Debug, Serialize, Deserialize)]
pub struct ConfigResponse {
    pub agent: AgentConfig,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct AgentConfig {
    pub model: String,
    pub temperature: f32,
}

pub async fn get_config(
    State(_state): State<AppState>,
) -> Result<Json<ConfigResponse>, StatusCode> {
    Ok(Json(ConfigResponse {
        agent: AgentConfig {
            model: "gpt-4".to_string(),
            temperature: 0.7,
        },
    }))
}

pub async fn update_config(
    State(_state): State<AppState>,
    Json(_config): Json<serde_json::Value>,
) -> Result<Json<ConfigResponse>, StatusCode> {
    Ok(Json(ConfigResponse {
        agent: AgentConfig {
            model: "gpt-4".to_string(),
            temperature: 0.7,
        },
    }))
}

pub async fn websocket(
    ws: WebSocket,
    State(_state): State<AppState>,
) {
    use axum::extract::ws::{Message, WebSocket};

    let (mut sender, mut receiver) = ws.split();

    // Handle incoming messages
    while let Some(msg) = receiver.next().await {
        match msg {
            Ok(Message::Text(text)) => {
                tracing::debug!("WebSocket received: {}", text);
                
                // Echo back for now
                if sender.send(Message::Text("pong".to_string())).await.is_err() {
                    break;
                }
            }
            Ok(Message::Binary(data)) => {
                tracing::debug!("WebSocket received binary: {} bytes", data.len());
            }
            Ok(Message::Ping(data)) => {
                if sender.send(Message::Pong(data)).await.is_err() {
                    break;
                }
            }
            Ok(Message::Pong(_)) => {}
            Ok(Message::Close(reason)) => {
                tracing::debug!("WebSocket closed: {:?}", reason);
                break;
            }
            Err(e) => {
                tracing::error!("WebSocket error: {}", e);
                break;
            }
        }
    }
}
