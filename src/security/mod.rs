//! Security Module
//! 
//! Zero-trust security system with threat detection and access control.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Security configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SecurityConfig {
    /// Enable zero-trust mode
    pub zero_trust: bool,
    /// Enable E2E encryption
    pub e2e_encryption: bool,
    /// Enable threat detection
    pub threat_detection: bool,
    /// Enable audit logging
    pub audit_logging: bool,
    /// Session timeout in seconds
    pub session_timeout_secs: u64,
    /// Max login attempts
    pub max_login_attempts: u32,
    /// Password requirements
    pub password_requirements: PasswordRequirements,
    /// Rate limiting
    pub rate_limit: RateLimitConfig,
}

impl Default for SecurityConfig {
    fn default() -> Self {
        Self {
            zero_trust: true,
            e2e_encryption: true,
            threat_detection: true,
            audit_logging: true,
            session_timeout_secs: 3600,
            max_login_attempts: 5,
            password_requirements: PasswordRequirements::default(),
            rate_limit: RateLimitConfig::default(),
        }
    }
}

/// Password requirements
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PasswordRequirements {
    pub min_length: u32,
    pub require_uppercase: bool,
    pub require_lowercase: bool,
    pub require_numbers: bool,
    pub require_special: bool,
}

impl Default for PasswordRequirements {
    fn default() -> Self {
        Self {
            min_length: 8,
            require_uppercase: true,
            require_lowercase: true,
            require_numbers: true,
            require_special: true,
        }
    }
}

/// Rate limit configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RateLimitConfig {
    pub enabled: bool,
    pub requests_per_minute: u32,
    pub burst_size: u32,
}

impl Default for RateLimitConfig {
    fn default() -> Self {
        Self {
            enabled: true,
            requests_per_minute: 60,
            burst_size: 10,
        }
    }
}

/// Trust level
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum TrustLevel {
    /// Fully isolated
    Untrusted = 0,
    /// Sandbox execution
    Low = 1,
    /// Limited permissions
    Medium = 2,
    /// Standard permissions
    High = 3,
    /// Fully trusted
    Full = 4,
}

impl TrustLevel {
    /// Get trust level from risk score
    pub fn from_risk_score(score: f64) -> Self {
        match score {
            s if s >= 0.9 => TrustLevel::Untrusted,
            s if s >= 0.7 => TrustLevel::Low,
            s if s >= 0.5 => TrustLevel::Medium,
            s if s >= 0.3 => TrustLevel::High,
            _ => TrustLevel::Full,
        }
    }
}

/// Permission mask
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct PermissionMask(u32);

impl PermissionMask {
    pub const READ: PermissionMask = PermissionMask(1 << 0);
    pub const WRITE: PermissionMask = PermissionMask(1 << 1);
    pub const EXECUTE: PermissionMask = PermissionMask(1 << 2);
    pub const ADMIN: PermissionMask = PermissionMask(1 << 3);
    pub const NETWORK: PermissionMask = PermissionMask(1 << 4);
    pub const FILESYSTEM: PermissionMask = PermissionMask(1 << 5);
    pub const PROCESS: PermissionMask = PermissionMask(1 << 6);
    pub const MEMORY: PermissionMask = PermissionMask(1 << 7);

    pub fn empty() -> Self {
        PermissionMask(0)
    }

    pub fn all() -> Self {
        PermissionMask(u32::MAX)
    }

    pub fn contains(&self, other: PermissionMask) -> bool {
        (self.0 & other.0) == other.0
    }
}

impl std::ops::BitOr for PermissionMask {
    type Output = Self;
    fn bitor(self, other: Self) -> Self {
        PermissionMask(self.0 | other.0)
    }
}

/// Identity type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum IdentityType {
    User,
    Bot,
    Service,
    Plugin,
    Anonymous,
}

/// Identity information
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Identity {
    pub id: String,
    pub identity_type: IdentityType,
    pub verified: bool,
    pub created_at: i64,
    pub last_seen: i64,
}

/// Trust assessment
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrustAssessment {
    pub identity_id: String,
    pub level: TrustLevel,
    pub risk_score: f64,
    pub factors: Vec<TrustFactor>,
    pub expires_at: Option<i64>,
}

/// Trust factor
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrustFactor {
    pub factor_type: TrustFactorType,
    pub weight: f64,
    pub value: f64,
    pub description: String,
}

/// Trust factor type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TrustFactorType {
    IdentityVerification,
    AuthenticationMethod,
    DeviceTrust,
    Location,
    BehaviorPattern,
    Reputation,
    History,
}

/// Threat type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ThreatType {
    Malware,
    Phishing,
    DataExfiltration,
    DenialOfService,
    PrivilegeEscalation,
    Injection,
    BruteForce,
    SocialEngineering,
    Unknown,
}

/// Severity level
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum Severity {
    Low = 1,
    Medium = 2,
    High = 3,
    Critical = 4,
}

/// Threat event
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThreatEvent {
    pub id: String,
    pub threat_type: ThreatType,
    pub severity: Severity,
    pub source: String,
    pub target: String,
    pub description: String,
    pub indicators: Vec<Indicator>,
    pub timestamp: i64,
    pub mitigated: bool,
}

/// Indicator of compromise
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Indicator {
    pub ioc_type: IOCType,
    pub value: String,
    pub confidence: f64,
}

/// IOC type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum IOCType {
    IP,
    Domain,
    URL,
    FileHash,
    Email,
    Registry,
    Process,
}

/// Security manager
pub struct SecurityManager {
    config: SecurityConfig,
    identities: RwLock<HashMap<String, Arc<Identity>>>,
    assessments: RwLock<HashMap<String, TrustAssessment>>,
    threat_events: RwLock<Vec<ThreatEvent>>,
}

impl SecurityManager {
    /// Create a new security manager
    pub fn new(config: &SecurityConfig) -> Self {
        Self {
            config: config.clone(),
            identities: RwLock::new(HashMap::new()),
            assessments: RwLock::new(HashMap::new()),
            threat_events: RwLock::new(Vec::new()),
        }
    }

    /// Assess trust for an identity
    pub async fn assess_trust(&self, identity_id: &str) -> Option<TrustAssessment> {
        let assessments = self.assessments.read().await;
        assessments.get(identity_id).cloned()
    }

    /// Check if an identity has required permissions
    pub async fn check_permission(&self, identity_id: &str, required: PermissionMask) -> bool {
        let assessments = self.assessments.read().await;
        if let Some(assessment) = assessments.get(identity_id) {
            let permissions = match assessment.level {
                TrustLevel::Untrusted => PermissionMask::empty(),
                TrustLevel::Low => PermissionMask::READ,
                TrustLevel::Medium => PermissionMask::READ | PermissionMask::EXECUTE,
                TrustLevel::High => PermissionMask::READ | PermissionMask::EXECUTE | PermissionMask::WRITE,
                TrustLevel::Full => PermissionMask::all(),
            };
            permissions.contains(required)
        } else {
            false
        }
    }

    /// Record a threat event
    pub async fn record_threat(&self, event: ThreatEvent) {
        let mut threats = self.threat_events.write().await;
        threats.push(event);
        
        // Keep only recent events
        if threats.len() > 10000 {
            threats.drain(0..1000);
        }
    }

    /// Get threat statistics
    pub async fn get_threat_stats(&self) -> ThreatStats {
        let threats = self.threat_events.read().await;
        
        let mut by_type: HashMap<ThreatType, u32> = HashMap::new();
        let mut by_severity: HashMap<Severity, u32> = HashMap::new();
        
        for event in threats.iter() {
            *by_type.entry(event.threat_type).or_insert(0) += 1;
            *by_severity.entry(event.severity).or_insert(0) += 1;
        }

        ThreatStats {
            total_events: threats.len(),
            by_type,
            by_severity,
            mitigated_count: threats.iter().filter(|e| e.mitigated).count(),
        }
    }

    /// Validate password
    pub fn validate_password(&self, password: &str) -> PasswordValidationResult {
        let reqs = &self.config.password_requirements;
        let mut errors = Vec::new();

        if password.len() < reqs.min_length as usize {
            errors.push(format!("Password must be at least {} characters", reqs.min_length));
        }
        if reqs.require_uppercase && !password.chars().any(|c| c.is_uppercase()) {
            errors.push("Password must contain at least one uppercase letter".to_string());
        }
        if reqs.require_lowercase && !password.chars().any(|c| c.is_lowercase()) {
            errors.push("Password must contain at least one lowercase letter".to_string());
        }
        if reqs.require_numbers && !password.chars().any(|c| c.is_numeric()) {
            errors.push("Password must contain at least one number".to_string());
        }
        if reqs.require_special && !password.chars().any(|c| !c.is_alphanumeric()) {
            errors.push("Password must contain at least one special character".to_string());
        }

        PasswordValidationResult {
            valid: errors.is_empty(),
            errors,
        }
    }

    /// Get configuration
    pub fn config(&self) -> &SecurityConfig {
        &self.config
    }
}

/// Threat statistics
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThreatStats {
    pub total_events: usize,
    pub by_type: HashMap<ThreatType, u32>,
    pub by_severity: HashMap<Severity, u32>,
    pub mitigated_count: usize,
}

/// Password validation result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PasswordValidationResult {
    pub valid: bool,
    pub errors: Vec<String>,
}
