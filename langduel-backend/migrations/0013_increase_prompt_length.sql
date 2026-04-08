-- Migration: Increase prompt field size in ai_phrases

ALTER TABLE ai_phrases ALTER COLUMN prompt TYPE VARCHAR(255);
