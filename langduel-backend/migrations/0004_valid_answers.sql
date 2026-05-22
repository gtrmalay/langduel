ALTER TABLE game_rounds ADD COLUMN IF NOT EXISTS valid_answers TEXT[] DEFAULT '{}';
