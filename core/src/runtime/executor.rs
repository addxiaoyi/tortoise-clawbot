//! Runtime executor - Task execution engine

use std::sync::Arc;
use tokio::sync::RwLock;
use std::future::Future;
use std::pin::Pin;
use crate::Error;

/// Task priority levels
#[derive(Debug, Clone, Copy, PartialEq, Eq, PartialOrd, Ord)]
pub enum Priority {
    Low = 0,
    Normal = 1,
    High = 2,
    Critical = 3,
}

/// Task definition
pub struct Task {
    pub id: String,
    pub name: String,
    pub priority: Priority,
    pub timeout_ms: Option<u64>,
}

impl Task {
    pub fn new(name: String) -> Self {
        Self {
            id: uuid::Uuid::new_v4().to_string(),
            name,
            priority: Priority::Normal,
            timeout_ms: Some(30000), // 30s default
        }
    }
    
    pub fn with_priority(mut self, priority: Priority) -> Self {
        self.priority = priority;
        self
    }
    
    pub fn with_timeout(mut self, timeout_ms: u64) -> Self {
        self.timeout_ms = Some(timeout_ms);
        self
    }
}

/// Task result
pub struct TaskResult<T> {
    pub task_id: String,
    pub success: bool,
    pub data: Option<T>,
    pub error: Option<String>,
    pub execution_time_ms: u64,
}

/// Task executor
pub struct Executor {
    workers: usize,
    queue: Arc<RwLock<TaskQueue>>,
}

impl Executor {
    pub fn new(workers: usize) -> Self {
        Self {
            workers,
            queue: Arc::new(RwLock::new(TaskQueue::new())),
        }
    }
    
    pub async fn submit<F, T>(&self, task: Task, future: F) -> Result<TaskResult<T>, Error>
    where
        F: Future<Output = Result<T, Error>>,
    {
        let start = std::time::Instant::now();
        let task_id = task.id.clone();
        
        // Execute with timeout
        let result = if let Some(timeout) = task.timeout_ms {
            tokio::time::timeout(
                std::time::Duration::from_millis(timeout),
                future
            ).await
        } else {
            Ok(tokio::time::timeout(
                std::time::Duration::MAX,
                future
            ).await.unwrap())
        };
        
        let execution_time_ms = start.elapsed().as_millis() as u64;
        
        match result {
            Ok(Ok(data)) => Ok(TaskResult {
                task_id,
                success: true,
                data: Some(data),
                error: None,
                execution_time_ms,
            }),
            Ok(Err(e)) => Ok(TaskResult {
                task_id,
                success: false,
                data: None,
                error: Some(e.to_string()),
                execution_time_ms,
            }),
            Err(_) => Ok(TaskResult {
                task_id,
                success: false,
                data: None,
                error: Some("Task timeout".to_string()),
                execution_time_ms,
            }),
        }
    }
}

/// Priority queue implementation
pub struct TaskQueue {
    tasks: Vec<Task>,
}

impl TaskQueue {
    pub fn new() -> Self {
        Self { tasks: Vec::new() }
    }
    
    pub fn push(&mut self, task: Task) {
        self.tasks.push(task);
        self.tasks.sort_by(|a, b| b.priority.cmp(&a.priority));
    }
    
    pub fn pop(&mut self) -> Option<Task> {
        self.tasks.pop()
    }
    
    pub fn len(&self) -> usize {
        self.tasks.len()
    }
    
    pub fn is_empty(&self) -> bool {
        self.tasks.is_empty()
    }
}

impl Default for TaskQueue {
    fn default() -> Self {
        Self::new()
    }
}

/// Worker pool
pub struct WorkerPool {
    executors: Vec<Executor>,
    current: Arc<std::sync::atomic::AtomicUsize>,
}

impl WorkerPool {
    pub fn new(size: usize) -> Self {
        let executors = (0..size)
            .map(|_| Executor::new(1))
            .collect();
        
        Self {
            executors,
            current: Arc::new(std::sync::atomic::AtomicUsize::new(0)),
        }
    }
    
    pub fn execute<F, T>(&self, task: Task, future: F) -> Pin<Box<dyn Future<Output = Result<TaskResult<T>, Error>>>
    where
        F: Future<Output = Result<T, Error>> + Send + 'static,
        T: Send + 'static,
    {
        // Round-robin selection
        let idx = self.current.fetch_add(1, std::sync::atomic::Ordering::Relaxed) 
            % self.executors.len();
        
        let executor = &self.executors[idx];
        Box::pin(executor.submit(task, future))
    }
}
