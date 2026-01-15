//! Session management module

use std::sync::Arc;
use tokio::sync::RwLock;
use dashmap::DashMap;
use uuid::Uuid;
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

/// Session status
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
#[serde(rename_all = "lowercase")]
pub enum SessionStatus {
    Active,
    Idle,
    Archived,
}

/// Session represents a conversation session
#[derive(Debug, Clone)]
pub struct Session {
    pub id: String,
    pub user_id: Option<String>,
    pub status: SessionStatus,
    pub created_at: DateTime<Utc>,
    pub updated_at: DateTime<Utc>,
    pub last_active_at: DateTime<Utc>,
    pub message_count: usize,
    pub config: SessionConfig,
    pub metadata: SessionMetadata,
}

/// Session configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SessionConfig {
    pub model: String,
    pub temperature: f32,
    pub max_tokens: usize,
    pub system_prompt: Option<String>,
}

impl Default for SessionConfig {
    fn default() -> Self {
        Self {
            model: "gpt-4o".to_string(),
            temperature: 0.7,
            max_tokens: 4096,
            system_prompt: None,
        }
    }
}

/// Session metadata
#[derive(Debug, Clone, Default, Serialize, Deserialize)]
pub struct SessionMetadata {
    pub channel: Option<String>,
    pub tags: Vec<String>,
    pub custom: serde_json::Value,
}

/// Session manager
pub struct SessionManager {
    sessions: Arc<DashMap<String, Session>>,
    max_sessions: usize,
}

impl SessionManager {
    /// Create a new session manager
    pub async fn new(max_sessions: usize) -> Result<Self, super::Error> {
        Ok(Self {
            sessions: Arc::new(DashMap::new()),
            max_sessions,
        })
    }

    /// Create a new session
    pub fn create(&self, user_id: Option<String>, config: Option<SessionConfig>) -> Session {
        let now = Utc::now();
        let session = Session {
            id: format!("sess_{}", Uuid::new_v4()),
            user_id,
            status: SessionStatus::Active,
            created_at: now,
            updated_at: now,
            last_active_at: now,
            message_count: 0,
            config: config.unwrap_or_default(),
            metadata: SessionMetadata::default(),
        };
        
        self.sessions.insert(session.id.clone(), session.clone());
        session
    }

    /// Get a session by ID
    pub fn get(&self, id: &str) -> Option<Session> {
        self.sessions.get(id).map(|s| s.clone())
    }

    /// Update a session
    pub fn update(&self, session: &Session) {
        self.sessions.insert(session.id.clone(), session.clone());
    }

    /// Delete a session
    pub fn delete(&self, id: &str) -> bool {
        self.sessions.remove(id).is_some()
    }

    /// List sessions with optional filters
    pub fn list(&self, user_id: Option<&str>, status: Option<SessionStatus>) -> Vec<Session> {
        self.sessions
            .iter()
            .filter(|s| {
                let user_match = user_id.map(|u| s.user_id.as_deref() == Some(u)).unwrap_or(true);
                let status_match = status.map(|st| s.status == st).unwrap_or(true);
                user_match && status_match
            })
            .map(|s| s.clone())
            .collect()
    }

    /// Touch a session to update last_active_at
    pub fn touch(&self, id: &str) -> bool {
        if let Some(mut session) = self.sessions.get_mut(id) {
            session.last_active_at = Utc::now();
            return true;
        }
        false
    }

    /// Increment message count
    pub fn increment_message_count(&self, id: &str) -> bool {
        if let Some(mut session) = self.sessions.get_mut(id) {
            session.message_count += 1;
            session.last_active_at = Utc::now();
            session.updated_at = Utc::now();
            return true;
        }
        false
    }

    /// Archive idle sessions
    pub fn archive_idle(&self, idle_duration_secs: i64) -> Vec<String> {
        let cutoff = Utc::now() - chrono::Duration::seconds(idle_duration_secs);
        let mut archived = Vec::new();
        
        for mut entry in self.sessions.iter_mut() {
            if entry.status == SessionStatus::Active && entry.last_active_at < cutoff {
                entry.status = SessionStatus::Idle;
                archived.push(entry.id.clone());
            }
        }
        
        archived
    }

    /// Get session statistics
    pub fn stats(&self) -> SessionStats {
        let total = self.sessions.len();
        let active = self.sessions
            .iter()
            .filter(|s| s.status == SessionStatus::Active)
            .count();
        let idle = self.sessions
            .iter()
            .filter(|s| s.status == SessionStatus::Idle)
            .count();
        
        SessionStats { total, active, idle }
    }

    /// Check if session limit is reached
    pub fn is_full(&self) -> bool {
        self.sessions.len() >= self.max_sessions
    }

    /// Get current session count
    pub fn len(&self) -> usize {
        self.sessions.len()
    }
}

/// Session statistics
#[derive(Debug, Clone)]
pub struct SessionStats {
    pub total: usize,
    pub active: usize,
    pub idle: usize,
}

impl Default for Session {
    fn default() -> Self {
        Self {
            id: String::new(),
            user_id: None,
            status: SessionStatus::Active,
            created_at: Utc::now(),
            updated_at: Utc::now(),
            last_active_at: Utc::now(),
            message_count: 0,
            config: SessionConfig::default(),
            metadata: SessionMetadata::default(),
        }
    }
}
