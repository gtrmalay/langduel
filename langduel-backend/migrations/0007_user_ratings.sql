-- user_ratings: ELO и ранги игроков
CREATE TABLE user_ratings (
  user_id UUID PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
  elo INT NOT NULL DEFAULT 1000,
  rank VARCHAR(20) NOT NULL DEFAULT 'newbie',
  games_played INT DEFAULT 0,
  wins INT DEFAULT 0,
  losses INT DEFAULT 0,
  current_streak INT DEFAULT 0,
  best_streak INT DEFAULT 0,
  total_losses_at_zero INT DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Индекс для быстрой сортировки лидерборда
CREATE INDEX idx_ratings_elo ON user_ratings(elo DESC);
CREATE INDEX idx_ratings_rank ON user_ratings(rank);
