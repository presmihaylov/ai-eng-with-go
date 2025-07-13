CREATE TABLE IF NOT EXISTS gocourse.agent_memory (
    id VARCHAR PRIMARY KEY,
    memory_content TEXT DEFAULT '',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Insert the default agent memory record
INSERT INTO gocourse.agent_memory (id, memory_content) 
VALUES ('agent', '') 
ON CONFLICT (id) DO NOTHING;