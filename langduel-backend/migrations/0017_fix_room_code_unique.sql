-- Allow multiple duels per room_code (rematches).
-- Previously UNIQUE(room_code) prevented creating a new duel after rematch,
-- and caused new games to reuse old finished duels, mixing analysis data.

ALTER TABLE duels DROP CONSTRAINT IF EXISTS duels_room_code_key;

CREATE INDEX IF NOT EXISTS idx_duels_room_code ON duels(room_code);
