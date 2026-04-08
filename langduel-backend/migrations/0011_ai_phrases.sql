-- Migration: Add AI phrases table for storing AI-generated translation phrases

CREATE TABLE IF NOT EXISTS ai_phrases (
    phrase_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    duel_id UUID REFERENCES duels(duel_id) ON DELETE CASCADE,
    prompt VARCHAR(255) NOT NULL,
    answers TEXT[] NOT NULL,
    topic VARCHAR(30) NOT NULL,
    difficulty VARCHAR(20) NOT NULL DEFAULT 'intermediate',
    lang_from VARCHAR(10) NOT NULL,
    lang_to VARCHAR(10) NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_ai_phrases_duel_id ON ai_phrases(duel_id);
CREATE INDEX IF NOT EXISTS idx_ai_phrases_used ON ai_phrases(used);
