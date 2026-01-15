//! Tortoise API Server
//! 
//! REST and WebSocket API server for Tortoise.

use anyhow::Result;
use axum::{
    extract::{Path, Query, State, WebSocket},
    http::StatusCode,
    response::Json,
    routing::{get, post},
    Router,
};
use serde::{Deserialize, Serialize};
use std::net::SocketAddr;
use std::sync::Arc;
use tokio::sync::RwLock;
use tower_http::cors::{Any, CorsLayer};
use tower_http::trace::TraceLayer;
use tracing_subscriber::{layer::SubscriberExt, util::SubscriberInitExt};

mod handlers;

#[derive(Clone)]
struct AppState {
    tortoise: Arc<RwLock<Option<tortoise_core::Tortoise>>>,
}

#[tokio::main]
async fn main() -> Result<()> {
    // Initialize logging
    tracing_subscriber::registry()
        .with(
            tracing_subscriber::EnvFilter::try_from_default_env()
                .unwrap_or_else(|_| "tortoise_api=debug,tower_http=debug".into()),
        )
        .with(tracing_subscriber::fmt::layer())
        .init();

    tracing::info!("Starting Tortoise API Server...");

    let state = AppState {
        tortoise: Arc::new(RwLock::new(None)),
    };

    // CORS
    let cors = CorsLayer::new()
        .allow_origin(Any)
        .allow_methods(Any)
        .allow_headers(Any);

    // Routes
    let app = Router::new()
        .route("/health", get(handlers::health))
        .route("/api/chat", post(handlers::chat))
        .route("/api/memory", get(handlers::get_memory))
        .route("/api/memory", post(handlers::store_memory))
        .route("/api/memory/stats", get(handlers::memory_stats))
        .route("/api/status", get(handlers::status))
        .route("/api/config", get(handlers::get_config))
        .route("/api/config", post(handlers::update_config))
        .route("/ws", get(handlers::websocket))
        .layer(TraceLayer::new_for_http())
        .layer(cors)
        .with_state(state);

    let addr = SocketAddr::from(([0, 0, 0, 0], 18789));
    tracing::info!("Listening on {}", addr);

    let listener = tokio::net::TcpListener::bind(addr).await?;
    axum::serve(listener, app).await?;

    Ok(())
}
