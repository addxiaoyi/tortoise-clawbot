//! Multi-Agent System Module

use super::*;
use std::collections::HashMap;
use tokio::sync::RwLock;

/// Agent role types
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum AgentRole {
    /// Main controller - task decomposition and result aggregation
    Orchestrator,
    /// Professional agent - specialized in specific domains
    Specialist,
    /// Assistant agent - daily assistance
    Assistant,
    /// Critic agent - review and suggest improvements
    Critic,
    /// Research agent - deep research and analysis
    Researcher,
    /// Coding agent - code development and debugging
    Coder,
    /// Testing agent - testing and quality assurance
    Tester,
    /// Documentation agent - documentation writing
    Documenter,
}

impl AgentRole {
    /// Get system prompt for this role
    pub fn system_prompt(&self) -> &'static str {
        match self {
            AgentRole::Orchestrator => r#"You are the Orchestrator, responsible for:
- Breaking down complex tasks into subtasks
- Assigning tasks to specialized agents
- Collecting and synthesizing results
- Managing workflow and dependencies
- Ensuring quality and consistency

Think step by step and communicate clearly."#,
            AgentRole::Specialist => r#"You are a Specialist agent, focused on:
- Deep expertise in your domain
- High-quality, accurate outputs
- Thorough analysis and solutions
- Clear explanations of complex topics

You are the expert in your field."#,
            AgentRole::Assistant => r#"You are a helpful Assistant, providing:
- Quick and accurate responses
- Friendly and approachable tone
- Concise and actionable information
- Proactive suggestions

You make users' lives easier."#,
            AgentRole::Critic => r#"You are a Critical thinker, responsible for:
- Reviewing and critiquing work
- Identifying flaws and improvements
- Ensuring quality standards
- Providing constructive feedback
- Validating correctness

You are thorough and unbiased."#,
            AgentRole::Researcher => r#"You are a Research expert, focused on:
- Deep investigation and analysis
- Finding credible sources
- Comprehensive understanding
- Evidence-based conclusions
- Clear documentation of findings

You leave no stone unturned."#,
            AgentRole::Coder => r#"You are an Expert Coder, skilled in:
- Clean, maintainable code
- Best practices and design patterns
- Performance optimization
- Testing and debugging
- Code reviews

You write code that others want to maintain."#,
            AgentRole::Tester => r#"You are a Quality Assurance expert, responsible for:
- Comprehensive test coverage
- Finding edge cases
- Validating functionality
- Bug reporting with details
- Quality metrics

You ensure software quality."#,
            AgentRole::Documenter => r#"You are a Technical Writer, skilled in:
- Clear, concise documentation
- Proper formatting and structure
- Examples and tutorials
- API documentation
- User guides

You make complex things understandable."#,
        }
    }

    /// Get capabilities for this role
    pub fn capabilities(&self) -> Vec<String> {
        match self {
            AgentRole::Orchestrator => vec![
                "task_decomposition".to_string(),
                "workflow_planning".to_string(),
                "result_synthesis".to_string(),
            ],
            AgentRole::Specialist => vec![
                "domain_expertise".to_string(),
                "deep_analysis".to_string(),
            ],
            AgentRole::Assistant => vec![
                "general_help".to_string(),
                "quick_responses".to_string(),
            ],
            AgentRole::Critic => vec![
                "review".to_string(),
                "feedback".to_string(),
                "quality_assurance".to_string(),
            ],
            AgentRole::Researcher => vec![
                "web_search".to_string(),
                "data_analysis".to_string(),
                "source_verification".to_string(),
            ],
            AgentRole::Coder => vec![
                "code_generation".to_string(),
                "code_review".to_string(),
                "debugging".to_string(),
            ],
            AgentRole::Tester => vec![
                "test_generation".to_string(),
                "bug_reporting".to_string(),
                "quality_metrics".to_string(),
            ],
            AgentRole::Documenter => vec![
                "technical_writing".to_string(),
                "api_docs".to_string(),
                "tutorials".to_string(),
            ],
        }
    }
}

/// Agent instance in the multi-agent system
#[derive(Debug, Clone)]
pub struct AgentInstance {
    /// Unique ID
    pub id: String,
    /// Name
    pub name: String,
    /// Role
    pub role: AgentRole,
    /// Description
    pub description: String,
    /// Underlying agent
    pub agent: Arc<dyn Agent>,
    /// Status
    pub status: RwLock<AgentStatus>,
    /// Capabilities
    pub capabilities: Vec<String>,
}

impl AgentInstance {
    /// Create a new agent instance
    pub fn new(
        id: String,
        name: String,
        role: AgentRole,
        description: String,
        agent: Arc<dyn Agent>,
        capabilities: Vec<String>,
    ) -> Self {
        Self {
            id,
            name,
            role,
            description,
            agent,
            status: RwLock::new(AgentStatus::Idle),
            capabilities,
        }
    }

    /// Update status
    pub async fn set_status(&self, status: AgentStatus) {
        *self.status.write().await = status;
    }

    /// Get status
    pub async fn get_status(&self) -> AgentStatus {
        self.status.read().await.clone()
    }
}

/// Task structure
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Task {
    /// Unique ID
    pub id: String,
    /// Task description
    pub description: String,
    /// Assigned agent ID
    pub assigned_to: Option<String>,
    /// Task status
    pub status: TaskStatus,
    /// Dependencies on other tasks
    pub dependencies: Vec<String>,
    /// Result when completed
    pub result: Option<TaskResult>,
    /// Created timestamp
    pub created_at: i64,
    /// Completed timestamp
    pub completed_at: Option<i64>,
}

/// Task status
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum TaskStatus {
    Pending,
    InProgress,
    Completed,
    Failed,
    Cancelled,
}

/// Task result
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TaskResult {
    /// Success flag
    pub success: bool,
    /// Output content
    pub output: String,
    /// Artifacts produced
    pub artifacts: Vec<Artifact>,
    /// Metrics
    pub metrics: HashMap<String, f64>,
}

/// Artifact produced by a task
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Artifact {
    /// Name
    pub name: String,
    /// Type
    pub artifact_type: ArtifactType,
    /// Content
    pub content: String,
    /// File path (optional)
    pub path: Option<String>,
}

/// Artifact type
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum ArtifactType {
    Text,
    Code,
    Document,
    Image,
    Data,
    Other,
}

/// Multi-agent system
pub struct MultiAgentSystem {
    /// Registered agents
    agents: RwLock<HashMap<String, Arc<AgentInstance>>>,
    /// Tasks
    tasks: RwLock<HashMap<String, Task>>,
    /// Orchestrator agent
    orchestrator: Arc<AgentInstance>,
    /// Message bus
    message_bus: Arc<MessageBus>,
}

impl MultiAgentSystem {
    /// Create a new multi-agent system
    pub fn new(orchestrator: Arc<dyn Agent>) -> Self {
        let message_bus = Arc::new(MessageBus::new());

        let orchestrator_instance = Arc::new(AgentInstance::new(
            "orchestrator".to_string(),
            "Orchestrator".to_string(),
            AgentRole::Orchestrator,
            "Main coordinator for task delegation".to_string(),
            orchestrator,
            AgentRole::Orchestrator.capabilities(),
        ));

        Self {
            agents: RwLock::new(HashMap::new()),
            tasks: RwLock::new(HashMap::new()),
            orchestrator: orchestrator_instance,
            message_bus,
        }
    }

    /// Register an agent
    pub async fn register_agent(&self, agent: AgentInstance) {
        let mut agents = self.agents.write().await;
        agents.insert(agent.id.clone(), Arc::new(agent));
    }

    /// Unregister an agent
    pub async fn unregister_agent(&self, id: &str) {
        let mut agents = self.agents.write().await;
        agents.remove(id);
    }

    /// Create a task
    pub async fn create_task(&self, description: String) -> String {
        let id = uuid::Uuid::new_v4().to_string();
        let task = Task {
            id: id.clone(),
            description,
            assigned_to: None,
            status: TaskStatus::Pending,
            dependencies: vec![],
            result: None,
            created_at: chrono::Utc::now().timestamp(),
            completed_at: None,
        };

        let mut tasks = self.tasks.write().await;
        tasks.insert(id.clone(), task);
        id
    }

    /// Execute a complex task
    pub async fn execute_task(&self, description: String) -> Result<TaskResult> {
        // Create main task
        let task_id = self.create_task(description).await;

        // Update status
        {
            let mut tasks = self.tasks.write().await;
            if let Some(task) = tasks.get_mut(&task_id) {
                task.status = TaskStatus::InProgress;
            }
        }

        // Plan task
        let plan = self.plan_task(&task_id).await?;

        // Execute plan
        let results = self.execute_plan(plan).await?;

        // Synthesize results
        let final_result = self.synthesize_results(results).await?;

        // Update task status
        {
            let mut tasks = self.tasks.write().await;
            if let Some(task) = tasks.get_mut(&task_id) {
                task.status = TaskStatus::Completed;
                task.result = Some(final_result.clone());
                task.completed_at = Some(chrono::Utc::now().timestamp());
            }
        }

        Ok(final_result)
    }

    /// Plan a task
    async fn plan_task(&self, task_id: &str) -> Result<ExecutionPlan> {
        let tasks = self.tasks.read().await;
        let task = tasks.get(task_id)
            .ok_or_else(|| anyhow::anyhow!("Task not found"))?;

        // Build prompt for orchestrator
        let prompt = format!(
            r#"Analyze this task and create an execution plan:

Task: {}

Available agents:
{}

Create a plan with:
1. Subtasks to complete
2. Dependencies between subtasks
3. Best agent for each subtask
4. Estimated complexity

Return a JSON plan."#,
            task.description,
            self.list_available_agents().await
        );

        let messages = vec![
            MessageBuilder::system(AgentRole::Orchestrator.system_prompt()).build(),
            MessageBuilder::user(prompt).build(),
        ];

        // Simplified - in real implementation, would parse JSON response
        Ok(ExecutionPlan {
            task_id: task_id.to_string(),
            subtasks: vec![SubTask {
                id: uuid::Uuid::new_v4().to_string(),
                description: task.description.clone(),
                assigned_role: AgentRole::Specialist,
                dependencies: vec![],
            }],
        })
    }

    /// Execute a plan
    async fn execute_plan(&self, plan: ExecutionPlan) -> Result<Vec<SubTaskResult>> {
        let mut results = Vec::new();

        // Execute in dependency order
        for subtask in &plan.subtasks {
            let result = self.execute_subtask(subtask).await?;
            results.push(result);
        }

        Ok(results)
    }

    /// Execute a subtask
    async fn execute_subtask(&self, subtask: &SubTask) -> Result<SubTaskResult> {
        // Find suitable agent
        let agent = self.find_agent_for_role(subtask.assigned_role).await
            .ok_or_else(|| anyhow::anyhow!("No agent available for role: {:?}", subtask.assigned_role))?;

        // Build prompt
        let prompt = format!(
            "{}\n\nTask: {}",
            agent.role.system_prompt(),
            subtask.description
        );

        let messages = vec![
            MessageBuilder::system(prompt).build(),
        ];

        // Execute
        let response = agent.agent.chat(messages, ChatOptions::default()).await?;

        // Collect result
        let mut output = String::new();
        while let Some(event) = response.events.recv().await {
            if let AgentEvent::Thinking(content) = event {
                output.push_str(&content);
            }
        }

        Ok(SubTaskResult {
            subtask_id: subtask.id.clone(),
            success: true,
            output,
            artifacts: vec![],
        })
    }

    /// Find an agent for a role
    async fn find_agent_for_role(&self, role: AgentRole) -> Option<Arc<AgentInstance>> {
        let agents = self.agents.read().await;
        agents.values().find(|a| a.role == role).cloned()
    }

    /// List available agents
    async fn list_available_agents(&self) -> String {
        let agents = self.agents.read().await;
        let mut output = String::new();

        for agent in agents.values() {
            output.push_str(&format!(
                "- {} ({:?}): {}\n",
                agent.name, agent.role, agent.description
            ));
        }

        if output.is_empty() {
            output = "No additional agents registered.".to_string();
        }

        output
    }

    /// Synthesize results from subtasks
    async fn synthesize_results(&self, results: Vec<SubTaskResult>) -> Result<TaskResult> {
        let combined_output: String = results
            .iter()
            .map(|r| format!("[{}] {}\n", r.subtask_id, r.output))
            .collect();

        // In a full implementation, would send to orchestrator for synthesis
        Ok(TaskResult {
            success: true,
            output: combined_output,
            artifacts: results.into_iter().flat_map(|r| r.artifacts).collect(),
            metrics: HashMap::new(),
        })
    }
}

/// Execution plan
#[derive(Debug, Clone)]
pub struct ExecutionPlan {
    /// Main task ID
    pub task_id: String,
    /// Subtasks
    pub subtasks: Vec<SubTask>,
}

/// Subtask
#[derive(Debug, Clone)]
pub struct SubTask {
    /// Unique ID
    pub id: String,
    /// Description
    pub description: String,
    /// Assigned role
    pub assigned_role: AgentRole,
    /// Dependencies
    pub dependencies: Vec<String>,
}

/// Subtask result
#[derive(Debug, Clone)]
pub struct SubTaskResult {
    /// Subtask ID
    pub subtask_id: String,
    /// Success flag
    pub success: bool,
    /// Output
    pub output: String,
    /// Artifacts
    pub artifacts: Vec<Artifact>,
}

/// Message bus for agent communication
pub struct MessageBus {
    channels: RwLock<HashMap<String, flume::Sender<AgentMessage>>>,
}

impl MessageBus {
    /// Create a new message bus
    pub fn new() -> Self {
        Self {
            channels: RwLock::new(HashMap::new()),
        }
    }

    /// Subscribe to messages for an agent
    pub async fn subscribe(&self, agent_id: &str) -> flume::Receiver<AgentMessage> {
        let (tx, rx) = flume::bounded(100);
        let mut channels = self.channels.write().await;
        channels.insert(agent_id.to_string(), tx);
        rx
    }

    /// Publish a message to a specific agent
    pub async fn publish(&self, target: &str, message: AgentMessage) -> Result<()> {
        let channels = self.channels.read().await;
        if let Some(tx) = channels.get(target) {
            tx.send(message).await?;
        }
        Ok(())
    }

    /// Broadcast to all agents
    pub async fn broadcast(&self, message: AgentMessage) -> Result<()> {
        let channels = self.channels.read().await;
        for tx in channels.values() {
            let _ = tx.send(message.clone()).await;
        }
        Ok(())
    }
}

impl Default for MessageBus {
    fn default() -> Self {
        Self::new()
    }
}

/// Agent message types
#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type", rename_all = "snake_case")]
pub enum AgentMessage {
    /// Task assigned
    TaskAssigned {
        task_id: String,
        from: String,
        to: String,
    },
    /// Task completed
    TaskCompleted {
        task_id: String,
        result: String,
    },
    /// Need help
    NeedHelp {
        from: String,
        question: String,
    },
    /// Feedback
    Feedback {
        from: String,
        to: String,
        content: String,
    },
    /// Status update
    StatusUpdate {
        agent_id: String,
        status: AgentStatus,
    },
}
