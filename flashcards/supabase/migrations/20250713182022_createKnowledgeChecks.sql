CREATE TABLE IF NOT EXISTS gocourse.knowledge_checks (
    id SERIAL PRIMARY KEY,
    note_id INTEGER NOT NULL,
    line_number_start INTEGER NOT NULL,
    line_number_end INTEGER NOT NULL,
    state VARCHAR NOT NULL CHECK (state IN ('pending', 'completed')),
    user_score INTEGER CHECK (user_score >= 1 AND user_score <= 10),
    user_score_explanation TEXT,
    topic_summary TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    FOREIGN KEY (note_id) REFERENCES gocourse.notes(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_knowledge_checks_created_at ON gocourse.knowledge_checks(created_at);
CREATE INDEX IF NOT EXISTS idx_knowledge_checks_note_id ON gocourse.knowledge_checks(note_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_checks_state ON gocourse.knowledge_checks(state);