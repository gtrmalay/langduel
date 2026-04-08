-- Add coins, win_streak, and unlocked_avatars to users

ALTER TABLE users ADD COLUMN IF NOT EXISTS coins INTEGER DEFAULT 0;

ALTER TABLE users ADD COLUMN IF NOT EXISTS win_streak INTEGER DEFAULT 0;

ALTER TABLE users ADD COLUMN IF NOT EXISTS unlocked_avatars TEXT DEFAULT '["default"]';

ALTER TABLE users ADD COLUMN IF NOT EXISTS sounds_enabled BOOLEAN DEFAULT true;
