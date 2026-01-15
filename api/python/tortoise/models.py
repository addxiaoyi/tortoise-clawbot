"""
Tortoise Models

Data models for the Tortoise Python SDK.
"""

from dataclasses import dataclass, field
from enum import Enum
from typing import List, Optional, Dict, Any
from datetime import datetime


class ThinkMode(Enum):
    """Thinking mode for agent responses."""

    FAST = "fast"
    BALANCED = "balanced"
    DEEP = "deep"
    RESEARCH = "research"
    CREATIVE = "creative"

    @property
    def timeout_ms(self) -> int:
        """Get timeout in milliseconds."""
        return {
            ThinkMode.FAST: 100,
            ThinkMode.BALANCED: 500,
            ThinkMode.DEEP: 2000,
            ThinkMode.RESEARCH: 5000,
            ThinkMode.CREATIVE: 10000,
        }.get(self, 500)

    @property
    def temperature(self) -> float:
        """Get default temperature."""
        return {
            ThinkMode.FAST: 0.0,
            ThinkMode.BALANCED: 0.5,
            ThinkMode.DEEP: 0.7,
            ThinkMode.RESEARCH: 0.6,
            ThinkMode.CREATIVE: 1.0,
        }.get(self, 0.5)


class MessageRole(Enum):
    """Message role in conversation."""

    SYSTEM = "system"
    USER = "user"
    ASSISTANT = "assistant"
    TOOL = "tool"


@dataclass
class Message:
    """Chat message."""

    role: MessageRole
    content: str
    metadata: Dict[str, Any] = field(default_factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        return {
            "role": self.role.value,
            "content": self.content,
            "metadata": self.metadata,
        }

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "Message":
        """Create from dictionary."""
        return cls(
            role=MessageRole(data.get("role", "user")),
            content=data.get("content", ""),
            metadata=data.get("metadata", {}),
        )

    @classmethod
    def system(cls, content: str) -> "Message":
        """Create a system message."""
        return cls(role=MessageRole.SYSTEM, content=content)

    @classmethod
    def user(cls, content: str) -> "Message":
        """Create a user message."""
        return cls(role=MessageRole.USER, content=content)

    @classmethod
    def assistant(cls, content: str) -> "Message":
        """Create an assistant message."""
        return cls(role=MessageRole.ASSISTANT, content=content)


@dataclass
class ChatOptions:
    """Options for chat requests."""

    thinking_mode: ThinkMode = ThinkMode.BALANCED
    temperature: Optional[float] = None
    max_tokens: Optional[int] = None
    system_prompt: Optional[str] = None
    tools: Optional[List[Dict[str, Any]]] = None
    stop: Optional[List[str]] = None

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary."""
        result = {"thinking_mode": self.thinking_mode.value}
        if self.temperature is not None:
            result["temperature"] = self.temperature
        if self.max_tokens is not None:
            result["max_tokens"] = self.max_tokens
        if self.system_prompt is not None:
            result["system_prompt"] = self.system_prompt
        if self.tools is not None:
            result["tools"] = self.tools
        if self.stop is not None:
            result["stop"] = self.stop
        return result


@dataclass
class MemoryItem:
    """Memory item."""

    id: str
    content: str
    importance: float
    memory_type: str
    created_at: datetime
    last_accessed: datetime
    access_count: int

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "MemoryItem":
        """Create from dictionary."""
        return cls(
            id=data.get("id", ""),
            content=data.get("content", ""),
            importance=data.get("importance", 0.0),
            memory_type=data.get("memory_type", "short_term"),
            created_at=datetime.fromisoformat(data.get("created_at", "1970-01-01T00:00:00Z")),
            last_accessed=datetime.fromisoformat(data.get("last_accessed", "1970-01-01T00:00:00Z")),
            access_count=data.get("access_count", 0),
        )


@dataclass
class AgentStatus:
    """Agent status information."""

    state: str
    model: str
    uptime_seconds: int
    memory_stats: Dict[str, int]

    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> "AgentStatus":
        """Create from dictionary."""
        return cls(
            state=data.get("state", "unknown"),
            model=data.get("model", "unknown"),
            uptime_seconds=data.get("uptime_seconds", 0),
            memory_stats=data.get("memory_stats", {}),
        )
