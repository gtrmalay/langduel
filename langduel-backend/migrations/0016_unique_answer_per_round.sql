-- Ensure one answer per (round, participant) so ON CONFLICT DO NOTHING works.
-- Postgres has no "ALTER TABLE ... ADD UNIQUE IF NOT EXISTS"; use a unique index instead.
CREATE UNIQUE INDEX IF NOT EXISTS idx_player_answers_round_participant
ON player_answers (round_id, participant_id);
