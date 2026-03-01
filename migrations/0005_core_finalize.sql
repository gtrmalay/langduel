-- Core finalize: winner, indexes

ALTER TABLE duels
  ADD COLUMN IF NOT EXISTS winner_user_id UUID REFERENCES users(user_id);

CREATE INDEX IF NOT EXISTS idx_duels_created_at ON duels(created_at);
CREATE INDEX IF NOT EXISTS idx_duels_created_by ON duels(created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_duel_participants_user ON duel_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_game_rounds_duel ON game_rounds(duel_id);
CREATE INDEX IF NOT EXISTS idx_player_answers_round ON player_answers(round_id);
