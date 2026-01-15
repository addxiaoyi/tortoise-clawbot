//! Security 模块 - 安全系统
//!
//! 零信任架构，端到端加密，威胁检测

mod trust;
mod crypto;
mod audit;

pub use trust::*;
pub use crypto::*;
pub use audit::*;

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::sync::Arc;
use tokio::sync::RwLock;

/// 安全配置
#[derive(Debug, Clone)]
pub struct SecurityConfig {
    pub encryption_enabled: bool,
    pub threat_detection_enabled: bool,
    pub trust_check_enabled: bool,
    pub audit_enabled: bool,
    pub max_trust_score: f32,
    pub min_trust_score: f32,
}

impl Default for SecurityConfig {
    fn default() -> Self {
        Self {
            encryption_enabled: true,
            threat_detection_enabled: true,
            trust_check_enabled: true,
            audit_enabled: true,
            max_trust_score: 1.0,
            min_trust_score: 0.0,
        }
    }
}

/// 威胁等级
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ThreatLevel {
    None,
    Low,
    Medium,
    High,
    Critical,
}

impl ThreatLevel {
    pub fn as_str(&self) -> &'static str {
        match self {
            ThreatLevel::None => "none",
            ThreatLevel::Low => "low",
            ThreatLevel::Medium => "medium",
            ThreatLevel::High => "high",
            ThreatLevel::Critical => "critical",
        }
    }
}

/// 信任分数
#[derive(Debug, Clone, Copy, Serialize, Deserialize)]
pub struct TrustScore {
    pub score: f32,
    pub level: TrustLevel,
    pub last_updated: i64,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TrustLevel {
    Blocked,
    Untrusted,
    Neutral,
    Trusted,
    FullyTrusted,
}

impl TrustScore {
    pub fn new(score: f32) -> Self {
        let level = match score {
            s if s < 0.2 => TrustLevel::Blocked,
            s if s < 0.4 => TrustLevel::Untrusted,
            s if s < 0.7 => TrustLevel::Neutral,
            s if s < 0.9 => TrustLevel::Trusted,
            _ => TrustLevel::FullyTrusted,
        };
        
        Self {
            score,
            level,
            last_updated: chrono::Utc::now().timestamp(),
        }
    }
}

/// 安全管理器
pub struct SecurityManager {
    config: SecurityConfig,
    trust_manager: TrustManager,
    crypto: CryptoManager,
    audit_logger: AuditLogger,
}

impl SecurityManager {
    pub fn new(config: SecurityConfig) -> Result<Self> {
        Ok(Self {
            config,
            trust_manager: TrustManager::new(),
            crypto: CryptoManager::new()?,
            audit_logger: AuditLogger::new()?,
        })
    }

    /// 检查威胁等级
    pub async fn check_threat(&self, content: &str) -> ThreatLevel {
        if !self.config.threat_detection_enabled {
            return ThreatLevel::None;
        }
        
        // 简单检测
        let dangerous_patterns = vec![
            "rm -rf", "drop table", "delete from",
            "exec(", "eval(", "system(",
        ];
        
        let content_lower = content.to_lowercase();
        for pattern in &dangerous_patterns {
            if content_lower.contains(&pattern.to_lowercase()) {
                return ThreatLevel::High;
            }
        }
        
        ThreatLevel::None
    }

    /// 检查信任分数
    pub async fn check_trust(&self, user_id: &str) -> TrustScore {
        if !self.config.trust_check_enabled {
            return TrustScore::new(0.5);
        }
        
        self.trust_manager.get_score(user_id).await
    }

    /// 加密数据
    pub fn encrypt(&self, data: &[u8], key: &[u8]) -> Result<Vec<u8>> {
        if !self.config.encryption_enabled {
            return Ok(data.to_vec());
        }
        
        self.crypto.encrypt(data, key)
    }

    /// 解密数据
    pub fn decrypt(&self, data: &[u8], key: &[u8]) -> Result<Vec<u8>> {
        if !self.config.encryption_enabled {
            return Ok(data.to_vec());
        }
        
        self.crypto.decrypt(data, key)
    }

    /// 记录审计日志
    pub async fn audit(&self, event: AuditEvent) -> Result<()> {
        if !self.config.audit_enabled {
            return Ok(());
        }
        
        self.audit_logger.log(event).await
    }
}

/// 审计事件
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AuditEvent {
    pub timestamp: i64,
    pub event_type: String,
    pub user_id: Option<String>,
    pub resource: Option<String>,
    pub action: String,
    pub result: String,
    pub metadata: serde_json::Value,
}

impl AuditEvent {
    pub fn new(event_type: &str, action: &str, result: &str) -> Self {
        Self {
            timestamp: chrono::Utc::now().timestamp(),
            event_type: event_type.to_string(),
            user_id: None,
            resource: None,
            action: action.to_string(),
            result: result.to_string(),
            metadata: Default::default(),
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_threat_level() {
        assert_eq!(ThreatLevel::None.as_str(), "none");
        assert_eq!(ThreatLevel::Critical.as_str(), "critical");
    }

    #[test]
    fn test_trust_score() {
        let score = TrustScore::new(0.5);
        assert_eq!(score.level, TrustLevel::Neutral);
        
        let blocked = TrustScore::new(0.1);
        assert_eq!(blocked.level, TrustLevel::Blocked);
    }

    #[tokio::test]
    async fn test_security_manager() {
        let config = SecurityConfig::default();
        let manager = SecurityManager::new(config).unwrap();
        
        let threat = manager.check_threat("Hello world").await;
        assert_eq!(threat, ThreatLevel::None);
    }
}
