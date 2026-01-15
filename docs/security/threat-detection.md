# Tortoise 安全系统设计

## 概述

Tortoise 安全系统提供全面的威胁检测、数据加密、权限管理和审计能力。

## 安全架构

```
┌─────────────────────────────────────────────────────────────────┐
│                      Tortoise Security Layer                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                     Trust Manager                          │  │
│  │   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐          │  │
│  │   │ Identity│ │ Reputation│ │ Policy │ │ Audit  │          │  │
│  │   └────────┘ └────────┘ └────────┘ └────────┘          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Crypto Module                            │  │
│  │   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐          │  │
│  │   │  E2E   │ │ Signal │ │  MAC   │ │  Keys  │          │  │
│  │   │ Encrypt│ │Protocol│ │ Verify │ │Manage  │          │  │
│  │   └────────┘ └────────┘ └────────┘ └────────┘          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                   Threat Detection                          │  │
│  │   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐          │  │
│  │   │ Anomaly │ │ Pattern│ │  ML    │ │  IOC   │          │  │
│  │   │ Detect │ │ Match  │ │ Detect │ │ Check  │          │  │
│  │   └────────┘ └────────┘ └────────┘ └────────┘          │  │
│  └──────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    Access Control                         │  │
│  │   ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐          │  │
│  │   │  RBAC  │ │  ABAC  │ │Permit  │ │  Rate  │          │  │
│  │   │        │ │        │ │  Check │ │ Limit  │          │  │
│  │   └────────┘ └────────┘ └────────┘ └────────┘          │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

## 核心实现

### 信任管理系统

```rust
// src/security/trust.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// 信任级别
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum TrustLevel {
    /// 完全隔离
    Untrusted = 0,
    /// 沙盒执行
    Low = 1,
    /// 有限权限
    Medium = 2,
    /// 标准权限
    High = 3,
    /// 完全信任
    Full = 4,
}

impl TrustLevel {
    /// 从风险评分获取信任级别
    pub fn from_risk_score(score: f64) -> Self {
        match score {
            s if s >= 0.9 => TrustLevel::Untrusted,
            s if s >= 0.7 => TrustLevel::Low,
            s if s >= 0.5 => TrustLevel::Medium,
            s if s >= 0.3 => TrustLevel::High,
            _ => TrustLevel::Full,
        }
    }

    /// 获取权限掩码
    pub fn permissions(&self) -> PermissionMask {
        match self {
            TrustLevel::Untrusted => PermissionMask::empty(),
            TrustLevel::Low => PermissionMask::READ,
            TrustLevel::Medium => PermissionMask::READ | PermissionMask::EXECUTE,
            TrustLevel::High => PermissionMask::READ | PermissionMask::EXECUTE | PermissionMask::WRITE,
            TrustLevel::Full => PermissionMask::all(),
        }
    }
}

/// 权限掩码
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
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

/// 身份信息
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Identity {
    pub id: String,
    pub identity_type: IdentityType,
    pub verified: bool,
    pub created_at: i64,
    pub last_seen: i64,
}

/// 身份类型
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum IdentityType {
    User,
    Bot,
    Service,
    Plugin,
    Anonymous,
}

/// 信任评估
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Trust评估 {
    pub identity_id: String,
    pub level: TrustLevel,
    pub risk_score: f64,
    pub factors: Vec<TrustFactor>,
    pub expires_at: Option<i64>,
}

/// 信任因素
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TrustFactor {
    pub factor_type: TrustFactorType,
    pub weight: f64,
    pub value: f64,
    pub description: String,
}

/// 信任因素类型
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

/// 信任管理器
pub struct TrustManager {
    identities: RwLock<HashMap<String, Arc<Identity>>>,
    assessments: RwLock<HashMap<String, Trust评估>>,
    policies: RwLock<Vec<TrustPolicy>>,
}

impl TrustManager {
    pub fn new() -> Self {
        Self {
            identities: RwLock::new(HashMap::new()),
            assessments: RwLock::new(HashMap::new()),
            policies: RwLock::new(Vec::new()),
        }
    }

    /// 评估信任
    pub async fn assess_trust(&self, identity: &Identity) -> Trust评估 {
        let mut factors = Vec::new();
        let mut total_weight = 0.0;
        let mut weighted_sum = 0.0;

        // 身份验证因素
        if identity.verified {
            factors.push(TrustFactor {
                factor_type: TrustFactorType::IdentityVerification,
                weight: 0.3,
                value: 1.0,
                description: "Identity verified".to_string(),
            });
            weighted_sum += 0.3;
        }
        total_weight += 0.3;

        // 身份类型因素
        let type_value = match identity.identity_type {
            IdentityType::User => 0.8,
            IdentityType::Bot => 0.6,
            IdentityType::Service => 0.7,
            IdentityType::Plugin => 0.4,
            IdentityType::Anonymous => 0.2,
        };
        factors.push(TrustFactor {
            factor_type: TrustFactorType::BehaviorPattern,
            weight: 0.4,
            value: type_value,
            description: format!("{:?} identity", identity.identity_type),
        });
        weighted_sum += 0.4 * type_value;
        total_weight += 0.4;

        // 历史因素
        let history_score = self.calculate_history_score(identity).await;
        factors.push(TrustFactor {
            factor_type: TrustFactorType::History,
            weight: 0.3,
            value: history_score,
            description: "Historical behavior score".to_string(),
        });
        weighted_sum += 0.3 * history_score;
        total_weight += 0.3;

        let risk_score = if total_weight > 0.0 {
            1.0 - (weighted_sum / total_weight)
        } else {
            1.0
        };

        let level = TrustLevel::from_risk_score(risk_score);

        Trust评估 {
            identity_id: identity.id.clone(),
            level,
            risk_score,
            factors,
            expires_at: Some(chrono::Utc::now().timestamp() + 3600),
        }
    }

    async fn calculate_history_score(&self, identity: &Identity) -> f64 {
        // 基于历史行为计算分数
        let last_seen_hours = (chrono::Utc::now().timestamp() - identity.last_seen) / 3600;
        
        // 越久未活跃，分数越低
        match last_seen_hours {
            h if h < 24 => 1.0,
            h if h < 168 => 0.9,  // 1 week
            h if h < 720 => 0.7, // 1 month
            _ => 0.5,
        }
    }

    /// 检查权限
    pub async fn check_permission(
        &self,
        identity_id: &str,
        required: PermissionMask,
    ) -> Result<bool> {
        let assessments = self.assessments.read().await;
        
        if let Some(assessment) = assessments.get(identity_id) {
            Ok(assessment.level.permissions().contains(required))
        } else {
            Ok(false)
        }
    }

    /// 添加策略
    pub async fn add_policy(&self, policy: TrustPolicy) {
        let mut policies = self.policies.write().await;
        policies.push(policy);
    }

    /// 评估请求
    pub async fn evaluate_request(&self, request: &TrustRequest) -> Result<Trust评估> {
        // 查找或创建身份
        let identities = self.identities.read().await;
        let identity = identities.get(&request.identity_id)
            .cloned()
            .ok_or_else(|| anyhow::anyhow!("Identity not found"))?;
        drop(identities);

        let mut assessment = self.assess_trust(&identity).await;

        // 应用策略
        let policies = self.policies.read().await;
        for policy in policies.iter() {
            if policy.matches(&request) {
                assessment = policy.apply(assessment);
            }
        }

        // 更新缓存
        let mut assessments = self.assessments.write().await;
        assessments.insert(request.identity_id.clone(), assessment.clone());

        Ok(assessment)
    }
}

/// 信任请求
#[derive(Debug, Clone)]
pub struct TrustRequest {
    pub identity_id: String,
    pub action: String,
    pub resource: String,
    pub context: HashMap<String, String>,
}

/// 信任策略
#[derive(Debug, Clone)]
pub struct TrustPolicy {
    pub id: String,
    pub name: String,
    pub conditions: Vec<PolicyCondition>,
    pub action: PolicyAction,
}

impl TrustPolicy {
    pub fn matches(&self, request: &TrustRequest) -> bool {
        self.conditions.iter().all(|c| c.matches(request))
    }

    pub fn apply(&self, assessment: Trust评估) -> Trust评估 {
        match self.action {
            PolicyAction::Increase(level) => {
                Trust评估 {
                    level: std::cmp::max(assessment.level, level),
                    ..assessment
                }
            }
            PolicyAction::Decrease(level) => {
                Trust评估 {
                    level: std::cmp::min(assessment.level, level),
                    ..assessment
                }
            }
            PolicyAction::Set(level) => {
                Trust评估 {
                    level,
                    ..assessment
                }
            }
        }
    }
}

/// 策略条件
#[derive(Debug, Clone)]
pub enum PolicyCondition {
    IdentityType(IdentityType),
    Action(String),
    Resource(String),
    TimeRange(i64, i64),
    IpRange(String, String),
}

/// 策略动作
#[derive(Debug, Clone)]
pub enum PolicyAction {
    Increase(TrustLevel),
    Decrease(TrustLevel),
    Set(TrustLevel),
}
```

### 威胁检测系统

```rust
// src/security/threat_detection.rs

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// 威胁类型
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

/// 威胁严重级别
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord, Serialize, Deserialize)]
pub enum Severity {
    Low = 1,
    Medium = 2,
    High = 3,
    Critical = 4,
}

/// 威胁事件
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

/// 威胁指标
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Indicator {
    pub ioc_type: IOCType,
    pub value: String,
    pub confidence: f64,
}

/// IOC 类型
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

/// 威胁检测器
pub struct ThreatDetector {
    rules: RwLock<Vec<DetectionRule>>,
    patterns: RwLock<HashMap<String, String>>,
    ml_model: Option<Arc<MLDetector>>,
    event_history: RwLock<Vec<ThreatEvent>>,
}

impl ThreatDetector {
    pub fn new() -> Self {
        Self {
            rules: RwLock::new(Vec::new()),
            patterns: RwLock::new(HashMap::new()),
            ml_model: None,
            event_history: RwLock::new(Vec::new()),
        }
    }

    /// 添加检测规则
    pub async fn add_rule(&self, rule: DetectionRule) {
        let mut rules = self.rules.write().await;
        rules.push(rule);
    }

    /// 检测威胁
    pub async fn detect(&self, event: &Event) -> Vec<ThreatEvent> {
        let mut threats = Vec::new();

        // 基于规则检测
        let rule_threats = self.detect_by_rules(event).await;
        threats.extend(rule_threats);

        // 基于模式检测
        let pattern_threats = self.detect_by_patterns(event).await;
        threats.extend(pattern_threats);

        // ML 检测
        if let Some(ref ml) = self.ml_model {
            let ml_threats = ml.detect(event).await;
            threats.extend(ml_threats);
        }

        // 存储事件
        for threat in &threats {
            let mut history = self.event_history.write().await;
            history.push(threat.clone());
            
            // 保留最近 10000 条
            if history.len() > 10000 {
                history.drain(0..1000);
            }
        }

        threats
    }

    async fn detect_by_rules(&self, event: &Event) -> Vec<ThreatEvent> {
        let mut threats = Vec::new();
        let rules = self.rules.read().await;

        for rule in rules.iter() {
            if rule.matches(event) {
                threats.push(ThreatEvent {
                    id: uuid::Uuid::new_v4().to_string(),
                    threat_type: rule.threat_type,
                    severity: rule.severity,
                    source: event.source.clone(),
                    target: event.target.clone(),
                    description: rule.description.clone(),
                    indicators: rule.indicators.clone(),
                    timestamp: chrono::Utc::now().timestamp(),
                    mitigated: false,
                });
            }
        }

        threats
    }

    async fn detect_by_patterns(&self, event: &Event) -> Vec<ThreatEvent> {
        let patterns = self.patterns.read().await;
        let mut threats = Vec::new();

        for (pattern_name, pattern) in patterns.iter() {
            if event.data.contains(pattern) {
                threats.push(ThreatEvent {
                    id: uuid::Uuid::new_v4().to_string(),
                    threat_type: ThreatType::Injection,
                    severity: Severity::High,
                    source: event.source.clone(),
                    target: event.target.clone(),
                    description: format!("Pattern match: {}", pattern_name),
                    indicators: vec![
                        Indicator {
                            ioc_type: IOCType::Process,
                            value: pattern.clone(),
                            confidence: 0.9,
                        }
                    ],
                    timestamp: chrono::Utc::now().timestamp(),
                    mitigated: false,
                });
            }
        }

        threats
    }

    /// 获取威胁统计
    pub async fn get_stats(&self) -> ThreatStats {
        let history = self.event_history.read().await;
        
        let mut by_type: HashMap<ThreatType, u32> = HashMap::new();
        let mut by_severity: HashMap<Severity, u32> = HashMap::new();
        
        for event in history.iter() {
            *by_type.entry(event.threat_type).or_insert(0) += 1;
            *by_severity.entry(event.severity).or_insert(0) += 1;
        }

        ThreatStats {
            total_events: history.len(),
            by_type,
            by_severity,
            mitigated_count: history.iter().filter(|e| e.mitigated).count(),
        }
    }
}

/// 检测规则
#[derive(Debug, Clone)]
pub struct DetectionRule {
    pub id: String,
    pub name: String,
    pub threat_type: ThreatType,
    pub severity: Severity,
    pub description: String,
    pub conditions: Vec<RuleCondition>,
    pub indicators: Vec<Indicator>,
}

impl DetectionRule {
    pub fn matches(&self, event: &Event) -> bool {
        self.conditions.iter().all(|c| c.matches(event))
    }
}

/// 规则条件
#[derive(Debug, Clone)]
pub enum RuleCondition {
    FieldEquals(String, String),
    FieldContains(String, String),
    FieldRegex(String, String),
    FieldIn(String, Vec<String>),
    ThresholdExceeded(String, u32),
    CountInWindow(String, u32, u64),
}

/// 事件
#[derive(Debug, Clone)]
pub struct Event {
    pub event_type: String,
    pub source: String,
    pub target: String,
    pub data: String,
    pub timestamp: i64,
}

/// 威胁统计
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ThreatStats {
    pub total_events: usize,
    pub by_type: HashMap<ThreatType, u32>,
    pub by_severity: HashMap<Severity, u32>,
    pub mitigated_count: usize,
}

/// ML 检测器
pub struct MLDetector {
    model: Arc<NeuralNetwork>,
    threshold: f64,
}

impl