//! Skill Module
//! 
//! Skill system for extending agent capabilities.

use anyhow::Result;
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::sync::Arc;
use tokio::sync::RwLock;

/// Skill configuration
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillConfig {
    /// Skill directory
    pub skill_dir: String,
    /// Enabled skills
    pub enabled_skills: Vec<String>,
    /// Auto-load skills
    pub auto_load: bool,
}

impl Default for SkillConfig {
    fn default() -> Self {
        Self {
            skill_dir: "skills".to_string(),
            enabled_skills: vec![],
            auto_load: true,
        }
    }
}

/// Skill metadata
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillMetadata {
    /// Skill ID
    pub id: String,
    /// Skill name
    pub name: String,
    /// Version
    pub version: String,
    /// Description
    pub description: String,
    /// Author
    pub author: String,
    /// Category
    pub category: SkillCategory,
    /// Tags
    pub tags: Vec<String>,
    /// Permissions required
    pub permissions: Vec<String>,
}

/// Skill category
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum SkillCategory {
    Utility,
    Productivity,
    Creative,
    Development,
    Research,
    Communication,
    Entertainment,
    System,
    Custom(String),
}

/// Skill status
#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum SkillStatus {
    Loaded,
    Enabled,
    Disabled,
    Error(String),
}

/// Skill instance
pub struct Skill {
    metadata: SkillMetadata,
    status: RwLock<SkillStatus>,
    executor: Box<dyn SkillExecutor>,
}

impl Skill {
    /// Create a new skill
    pub fn new(metadata: SkillMetadata, executor: Box<dyn SkillExecutor>) -> Self {
        Self {
            metadata,
            status: RwLock::new(SkillStatus::Loaded),
            executor,
        }
    }

    /// Execute the skill
    pub async fn execute(&self, context: SkillContext) -> Result<SkillResult> {
        *self.status.write().await = SkillStatus::Enabled;
        let result = self.executor.execute(context).await;
        *self.status.write().await = SkillStatus::Loaded;
        result
    }

    /// Get skill metadata
    pub fn metadata(&self) -> &SkillMetadata {
        &self.metadata
    }

    /// Get skill status
    pub async fn status(&self) -> SkillStatus {
        self.status.read().await.clone()
    }
}

/// Skill executor trait
#[async_trait::async_trait]
pub trait SkillExecutor: Send + Sync {
    /// Execute the skill
    async fn execute(&self, context: SkillContext) -> Result<SkillResult>;

    /// Validate input
    fn validate_input(&self, input: &serde_json::Value) -> Result<()>;
}

/// Skill context
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillContext {
    /// Input parameters
    pub input: serde_json::Value,
    /// User ID
    pub user_id: Option<String>,
    /// Session ID
    pub session_id: Option<String>,
    /// Channel
    pub channel: Option<String>,
    /// Metadata
    pub metadata: HashMap<String, serde_json::Value>,
}

/// Skill result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillResult {
    /// Success flag
    pub success: bool,
    /// Output
    pub output: serde_json::Value,
    /// Error message (if failed)
    pub error: Option<String>,
    /// Artifacts
    pub artifacts: Vec<SkillArtifact>,
    /// Execution time in ms
    pub execution_time_ms: u64,
}

/// Skill artifact
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct SkillArtifact {
    /// Artifact name
    pub name: String,
    /// Artifact type
    pub artifact_type: String,
    /// Content
    pub content: String,
    /// MIME type
    pub mime_type: Option<String>,
}

/// Skill manager
pub struct SkillManager {
    skills: RwLock<HashMap<String, Arc<Skill>>>,
    config: SkillConfig,
}

impl SkillManager {
    /// Create a new skill manager
    pub fn new(config: SkillConfig) -> Self {
        Self {
            skills: RwLock::new(HashMap::new()),
            config,
        }
    }

    /// Register a skill
    pub async fn register_skill(&self, skill: Arc<Skill>) {
        let mut skills = self.skills.write().await;
        skills.insert(skill.metadata.id.clone(), skill);
    }

    /// Get a skill by ID
    pub async fn get_skill(&self, id: &str) -> Option<Arc<Skill>> {
        let skills = self.skills.read().await;
        skills.get(id).cloned()
    }

    /// List all skills
    pub async fn list_skills(&self) -> Vec<SkillMetadata> {
        let skills = self.skills.read().await;
        skills.values()
            .map(|s| s.metadata().clone())
            .collect()
    }

    /// Execute a skill
    pub async fn execute_skill(&self, id: &str, context: SkillContext) -> Result<SkillResult> {
        let skill = self.get_skill(id).await
            .ok_or_else(|| anyhow::anyhow!("Skill not found: {}", id))?;
        
        skill.execute(context).await
    }

    /// Enable a skill
    pub async fn enable_skill(&self, id: &str) -> Result<()> {
        let mut skills = self.skills.write().await;
        if let Some(skill) = skills.get_mut(id) {
            *skill.status.write().await = SkillStatus::Enabled;
        }
        Ok(())
    }

    /// Disable a skill
    pub async fn disable_skill(&self, id: &str) -> Result<()> {
        let mut skills = self.skills.write().await;
        if let Some(skill) = skills.get_mut(id) {
            *skill.status.write().await = SkillStatus::Disabled;
        }
        Ok(())
    }
}

/// Built-in skill implementations

/// Calculator skill
pub struct CalculatorSkill;

impl CalculatorSkill {
    pub fn new() -> Self {
        Self
    }
}

#[async_trait::async_trait]
impl SkillExecutor for CalculatorSkill {
    async fn execute(&self, context: SkillContext) -> Result<SkillResult> {
        let start = std::time::Instant::now();
        
        let expression = context.input.get("expression")
            .and_then(|v| v.as_str())
            .unwrap_or("");

        // Simple expression evaluation (in production, use proper parser)
        let result = eval_simple_expr(expression);

        Ok(SkillResult {
            success: true,
            output: serde_json::json!({
                "expression": expression,
                "result": result
            }),
            error: None,
            artifacts: vec![],
            execution_time_ms: start.elapsed().as_millis() as u64,
        })
    }

    fn validate_input(&self, input: &serde_json::Value) -> Result<()> {
        if !input.get("expression").is_some() {
            anyhow::bail!("Missing 'expression' field");
        }
        Ok(())
    }
}

fn eval_simple_expr(expr: &str) -> f64 {
    // Very simplified - just return 0
    // In production, use proper math expression parser
    tracing::debug!("Evaluating expression: {}", expr);
    0.0
}

/// Web search skill
pub struct WebSearchSkill;

impl WebSearchSkill {
    pub fn new() -> Self {
        Self
    }
}

#[async_trait::async_trait]
impl SkillExecutor for WebSearchSkill {
    async fn execute(&self, context: SkillContext) -> Result<SkillResult> {
        let start = std::time::Instant::now();
        
        let query = context.input.get("query")
            .and_then(|v| v.as_str())
            .unwrap_or("");

        // Placeholder - in production, call actual search API
        let results = vec![
            serde_json::json!({
                "title": "Search result 1",
                "url": "https://example.com/1",
                "snippet": "This is a search result"
            })
        ];

        Ok(SkillResult {
            success: true,
            output: serde_json::json!({
                "query": query,
                "results": results
            }),
            error: None,
            artifacts: vec![],
            execution_time_ms: start.elapsed().as_millis() as u64,
        })
    }

    fn validate_input(&self, input: &serde_json::Value) -> Result<()> {
        if !input.get("query").is_some() {
            anyhow::bail!("Missing 'query' field");
        }
        Ok(())
    }
}
