-- Tortoise Database Schema
-- PostgreSQL 16+

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================
-- Sessions Table
-- ============================================
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    config JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    message_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_active_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_status CHECK (status IN ('active', 'idle', 'archived'))
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_status ON sessions(status);
CREATE INDEX idx_sessions_last_active ON sessions(last_active_at);
CREATE INDEX idx_sessions_created ON sessions(created_at);

-- ============================================
-- Messages Table
-- ============================================
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'text',
    attachments JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT valid_role CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE INDEX idx_messages_session ON messages(session_id);
CREATE INDEX idx_messages_created ON messages(created_at);
CREATE INDEX idx_messages_role ON messages(role);

-- ============================================
-- Memory Table
-- ============================================
CREATE TABLE memory (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    user_id VARCHAR(255),
    content TEXT NOT NULL,
    memory_type VARCHAR(50) NOT NULL DEFAULT 'semantic',
    importance REAL NOT NULL DEFAULT 0.5,
    tags TEXT[] DEFAULT '{}',
    vector_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    accessed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    access_count INTEGER NOT NULL DEFAULT 0,
    
    CONSTRAINT valid_memory_type CHECK (memory_type IN ('working', 'semantic', 'episodic')),
    CONSTRAINT valid_importance CHECK (importance >= 0 AND importance <= 1)
);

CREATE INDEX idx_memory_session ON memory(session_id);
CREATE INDEX idx_memory_user ON memory(user_id);
CREATE INDEX idx_memory_type ON memory(memory_type);
CREATE INDEX idx_memory_created ON memory(created_at);
CREATE INDEX idx_memory_accessed ON memory(accessed_at);
CREATE INDEX idx_memory_tags ON memory USING GIN(tags);

-- ============================================
-- Plugins Table
-- ============================================
CREATE TABLE plugins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author VARCHAR(255),
    enabled BOOLEAN NOT NULL DEFAULT true,
    config JSONB DEFAULT '{}',
    dependencies TEXT[] DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_plugins_enabled ON plugins(enabled);

-- ============================================
-- Tools Table
-- ============================================
CREATE TABLE tools (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    parameters_schema JSONB NOT NULL DEFAULT '{}',
    enabled BOOLEAN NOT NULL DEFAULT true,
    plugin_id UUID REFERENCES plugins(id) ON DELETE SET NULL,
    usage_count INTEGER NOT NULL DEFAULT 0,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tools_name ON tools(name);
CREATE INDEX idx_tools_enabled ON tools(enabled);
CREATE INDEX idx_tools_plugin ON tools(plugin_id);

-- ============================================
-- Tool Executions Table (Audit Log)
-- ============================================
CREATE TABLE tool_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id UUID REFERENCES sessions(id) ON DELETE SET NULL,
    tool_id UUID REFERENCES tools(id) ON DELETE SET NULL,
    tool_name VARCHAR(255) NOT NULL,
    arguments JSONB NOT NULL DEFAULT '{}',
    result JSONB,
    success BOOLEAN NOT NULL,
    error TEXT,
    execution_time_ms INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_tool_executions_session ON tool_executions(session_id);
CREATE INDEX idx_tool_executions_tool ON tool_executions(tool_name);
CREATE INDEX idx_tool_executions_created ON tool_executions(created_at);

-- ============================================
-- API Keys Table
-- ============================================
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    user_id VARCHAR(255),
    name VARCHAR(255),
    rate_limit INTEGER NOT NULL DEFAULT 60,
    enabled BOOLEAN NOT NULL DEFAULT true,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_api_keys_hash ON api_keys(key_hash);
CREATE INDEX idx_api_keys_user ON api_keys(user_id);
CREATE INDEX idx_api_keys_enabled ON api_keys(enabled);

-- ============================================
-- Audit Log Table
-- ============================================
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id VARCHAR(255),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(100),
    resource_id UUID,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_user ON audit_log(user_id);
CREATE INDEX idx_audit_action ON audit_log(action);
CREATE INDEX idx_audit_resource ON audit_log(resource_type, resource_id);
CREATE INDEX idx_audit_created ON audit_log(created_at);

-- ============================================
-- Functions and Triggers
-- ============================================

-- Update updated_at trigger
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_sessions_updated
    BEFORE UPDATE ON sessions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trigger_plugins_updated
    BEFORE UPDATE ON plugins
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- Increment message count on insert
CREATE OR REPLACE FUNCTION increment_message_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE sessions SET message_count = message_count + 1 WHERE id = NEW.session_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_message_inserted
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION increment_message_count();

-- Update last_active_at on new message
CREATE OR REPLACE FUNCTION touch_session()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE sessions SET last_active_at = NOW() WHERE id = NEW.session_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_session_touched
    AFTER INSERT ON messages
    FOR EACH ROW
    EXECUTE FUNCTION touch_session();

-- Archive idle sessions (runs daily)
CREATE OR REPLACE FUNCTION archive_idle_sessions()
RETURNS INTEGER AS $$
DECLARE
    archived_count INTEGER;
BEGIN
    WITH updated AS (
        UPDATE sessions 
        SET status = 'idle'
        WHERE status = 'active' 
        AND last_active_at < NOW() - INTERVAL '7 days'
        RETURNING id
    )
    SELECT COUNT(*) INTO archived_count FROM updated;
    RETURN archived_count;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Views
-- ============================================

-- Active sessions view
CREATE VIEW v_active_sessions AS
SELECT 
    s.*,
    COUNT(m.id) AS message_count,
    MAX(m.created_at) AS last_message_at
FROM sessions s
LEFT JOIN messages m ON s.id = m.session_id
WHERE s.status = 'active'
GROUP BY s.id;

-- Popular tools view
CREATE VIEW v_popular_tools AS
SELECT 
    t.name,
    t.description,
    COUNT(e.id) AS execution_count,
    AVG(e.execution_time_ms) AS avg_execution_time
FROM tools t
LEFT JOIN tool_executions e ON t.id = e.tool_id
GROUP BY t.id;

-- Memory statistics view
CREATE VIEW v_memory_stats AS
SELECT 
    memory_type,
    COUNT(*) AS count,
    AVG(importance) AS avg_importance,
    COUNT(*) FILTER (WHERE access_count > 0) AS accessed_count
FROM memory
GROUP BY memory_type;
