-- Add coins to user_ratings and coins_reward to achievements
ALTER TABLE user_ratings ADD COLUMN IF NOT EXISTS coins INT DEFAULT 0;

ALTER TABLE achievements ADD COLUMN IF NOT EXISTS coins_reward INT DEFAULT 0;

-- Update existing achievements with coins rewards
UPDATE achievements SET coins_reward = 5 WHERE id = 'first_win';
UPDATE achievements SET coins_reward = 15 WHERE id = 'warrior';
UPDATE achievements SET coins_reward = 30 WHERE id = 'veteran';
UPDATE achievements SET coins_reward = 75 WHERE id = 'champion';
UPDATE achievements SET coins_reward = 10 WHERE id = 'streak_5';
UPDATE achievements SET coins_reward = 35 WHERE id = 'streak_10';
UPDATE achievements SET coins_reward = 5 WHERE id = 'games_10';
UPDATE achievements SET coins_reward = 20 WHERE id = 'games_50';
