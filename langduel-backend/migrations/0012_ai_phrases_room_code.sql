-- Migration: Add room_code to ai_phrases for easy lookup

ALTER TABLE ai_phrases ADD COLUMN IF NOT EXISTS room_code VARCHAR(50);
CREATE INDEX IF NOT EXISTS idx_ai_phrases_room_code ON ai_phrases(room_code);
