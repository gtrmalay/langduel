-- Add correct_answer column to game_rounds for analysis

ALTER TABLE game_rounds ADD COLUMN IF NOT EXISTS correct_answer VARCHAR(255);
