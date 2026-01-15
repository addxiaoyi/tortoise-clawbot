"""
Tortoise Models
Data classes for Tortoise API responses.
"""

from dataclasses import dataclass, field
from typing import Optional, Any
from datetime import datetime


@dataclass
class ChatMessage:
    """Chat message model."""
    id: str
    role: str
    content: str
    model: Optional[str] = None
    session_id: Optional[str] = None
    created_at: Optional[datetime] = None
    metadata: dict = field(default_factory=dict)
    
    @classmethod
    def from_dict(cls, data: dict) -> "ChatMessage":
        """Create from dictionary."""
        created_at = data.get("created_at")
        if created_at and isinstance(created_at, str):
            created_at = datetime.fromisoformat(created_at.replace("Z", "+00:00"))
        
        return cls(
            id=data.get("id", ""),
            role=data.get("role", "assistant"),
            content=data.get("content", ""),
            model=data.get("model"),
            session_id=data.get("session_id"),
            created_at=created_at,
            metadata=data.get("metadata", {}),
        )
    
    def to_dict(self) -> dict:
        """Convert to dictionary."""
        return {
            "id": self.id,
            "role": self.role,
            "content": self.content,
            "model": self.model,
            "session_id": self.session_id,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "metadata": self.metadata,
        }


@dataclass
class Session:
    """Chat session model."""
    id: str
    title: str
    model: Optional[str] = None
    messages: list[ChatMessage] = field(default_factory=list)
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    metadata: dict = field(default_factory=dict)
    
    @classmethod
    def from_dict(cls, data: dict) -> "Session":
        """Create from dictionary."""
        created_at = data.get("created_at")
        if created_at and isinstance(created_at, str):
            created_at = datetime.fromisoformat(created_at.replace("Z", "+00:00"))
        
        updated_at = data.get("updated_at")
        if updated_at and isinstance(updated_at, str):
            updated_at = datetime.fromisoformat(updated_at.replace("Z", "+00:00"))
        
        messages = [
            ChatMessage.from_dict(m) 
            for m in data.get("messages", [])
        ]
        
        return cls(
            id=data.get("id", ""),
            title=data.get("title", "Untitled"),
            model=data.get("model"),
            messages=messages,
            created_at=created_at,
            updated_at=updated_at,
            metadata=data.get("metadata", {}),
        )


@dataclass
class Plugin:
    """Plugin model."""
    id: str
    name: str
    version: str
    description: str = ""
    author: str = ""
    enabled: bool = False
    config: dict = field(default_factory=dict)
    
    @classmethod
    def from_dict(cls, data: dict) -> "Plugin":
        """Create from dictionary."""
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            version=data.get("version", ""),
            description=data.get("description", ""),
            author=data.get("author", ""),
            enabled=data.get("enabled", False),
            config=data.get("config", {}),
        )


@dataclass
class Channel:
    """Message channel model."""
    id: str
    name: str
    type: str
    enabled: bool = False
    status: str = "offline"
    config: dict = field(default_factory=dict)
    
    @classmethod
    def from_dict(cls, data: dict) -> "Channel":
        """Create from dictionary."""
        return cls(
            id=data.get("id", ""),
            name=data.get("name", ""),
            type=data.get("type", ""),
            enabled=data.get("enabled", False),
            status=data.get("status", "offline"),
            config=data.get("config", {}),
        )


@dataclass
class MemoryEntry:
    """Memory/knowledge base entry."""
    id: str
    content: str
    type: str = "fact"
    importance: float = 0.5
    access_count: int = 0
    last_accessed: Optional[datetime] = None
    created_at: Optional[datetime] = None
    metadata: dict = field(default_factory=dict)
    
    @classmethod
    def from_dict(cls, data: dict) -> "MemoryEntry":
        """Create from dictionary."""
        last_accessed = data.get("last_accessed")
        if last_accessed and isinstance(last_accessed, str):
            last_accessed = datetime.fromisoformat(last_accessed.replace("Z", "+00:00"))
        
        created_at = data.get("created_at")
        if created_at and isinstance(created_at, str):
            created_at = datetime.fromisoformat(created_at.replace("Z", "+00:00"))
        
        return cls(
            id=data.get("id", ""),
            content=data.get("content", ""),
            type=data.get("type", "fact"),
            importance=data.get("importance", 0.5),
            access_count=data.get("access_count", 0),
            last_accessed=last_accessed,
            created_at=created_at,
            metadata=data.get("metadata", {}),
        )
