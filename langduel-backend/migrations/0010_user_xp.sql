-- Add XP to user_ratings
ALTER TABLE user_ratings ADD COLUMN IF NOT EXISTS xp BIGINT DEFAULT 0;
