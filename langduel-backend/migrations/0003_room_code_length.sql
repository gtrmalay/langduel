-- Increase room_code length for generated IDs like "room-xxxxxx"
ALTER TABLE duels
  ALTER COLUMN room_code TYPE VARCHAR(32);
