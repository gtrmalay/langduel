CREATE TABLE IF NOT EXISTS user_statistics (
  stats_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
  total_duels_played INT NOT NULL DEFAULT 0,
  total_duels_won INT NOT NULL DEFAULT 0,
  total_score BIGINT NOT NULL DEFAULT 0,
  overall_accuracy FLOAT NOT NULL DEFAULT 0,
  best_win_streak INT NOT NULL DEFAULT 0,
  total_play_time_min INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
