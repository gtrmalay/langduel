-- LangDuel Full schema (PostgreSQL)
-- Extends MVP with achievements, skins, statistics.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(30) NOT NULL UNIQUE,
  email VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255),
  is_guest BOOLEAN NOT NULL DEFAULT FALSE,
  guest_expires_at TIMESTAMPTZ,
  avatar_url VARCHAR(255),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE phrases (
  phrase_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  text VARCHAR(255) NOT NULL,
  lang VARCHAR(10) NOT NULL,
  topic VARCHAR(30) NOT NULL,
  difficulty INT NOT NULL DEFAULT 1
);

CREATE TABLE duels (
  duel_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code VARCHAR(32) NOT NULL UNIQUE,
  created_by_user_id UUID NOT NULL REFERENCES users(user_id),
  winner_user_id UUID REFERENCES users(user_id),
  language_from VARCHAR(10) NOT NULL DEFAULT 'en',
  language_to VARCHAR(10) NOT NULL DEFAULT 'ru',
  theme VARCHAR(30) NOT NULL DEFAULT 'default',
  difficulty INT NOT NULL DEFAULT 1,
  max_rounds INT NOT NULL DEFAULT 10,
  status VARCHAR(20) NOT NULL DEFAULT 'waiting',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ
);

CREATE TABLE duel_participants (
  participant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  duel_id UUID NOT NULL REFERENCES duels(duel_id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(user_id),
  player_order INT NOT NULL,
  initial_hp INT NOT NULL DEFAULT 100,
  final_hp INT,
  score INT NOT NULL DEFAULT 0,
  joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (duel_id, user_id)
);

CREATE TABLE game_rounds (
  round_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  duel_id UUID NOT NULL REFERENCES duels(duel_id) ON DELETE CASCADE,
  round_number INT NOT NULL,
  phrase_id UUID NOT NULL REFERENCES phrases(phrase_id),
  time_limit_ms INT NOT NULL DEFAULT 10000,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (duel_id, round_number)
);

CREATE TABLE player_answers (
  answer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  round_id UUID NOT NULL REFERENCES game_rounds(round_id) ON DELETE CASCADE,
  participant_id UUID NOT NULL REFERENCES duel_participants(participant_id) ON DELETE CASCADE,
  translation_text TEXT NOT NULL,
  is_correct BOOLEAN NOT NULL DEFAULT FALSE,
  response_time_ms INT NOT NULL,
  damage_dealt INT NOT NULL DEFAULT 0,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE achievements (
  achievement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon_url VARCHAR(255),
  condition_type VARCHAR(30) NOT NULL,
  condition_value INT,
  is_secret BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE user_achievements (
  user_achievement_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  achievement_id UUID NOT NULL REFERENCES achievements(achievement_id) ON DELETE CASCADE,
  unlocked_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  progress_before_unlock INT NOT NULL DEFAULT 0,
  UNIQUE (user_id, achievement_id)
);

CREATE TABLE skins (
  skin_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name VARCHAR(50) NOT NULL,
  description TEXT,
  image_url VARCHAR(255),
  unlock_condition VARCHAR(30) NOT NULL,
  unlock_value VARCHAR(100),
  price_coins INT NOT NULL DEFAULT 0
);

CREATE TABLE user_skins (
  user_skin_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  skin_id UUID NOT NULL REFERENCES skins(skin_id) ON DELETE CASCADE,
  acquired_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  is_active BOOLEAN NOT NULL DEFAULT FALSE,
  UNIQUE (user_id, skin_id)
);

CREATE TABLE user_statistics (
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

CREATE TABLE language_statistics (
  lang_stats_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id) ON DELETE CASCADE,
  language_from VARCHAR(10) NOT NULL,
  language_to VARCHAR(10) NOT NULL,
  duels_played INT NOT NULL DEFAULT 0,
  duels_won INT NOT NULL DEFAULT 0,
  accuracy FLOAT NOT NULL DEFAULT 0,
  words_translated INT NOT NULL DEFAULT 0,
  UNIQUE (user_id, language_from, language_to)
);

CREATE INDEX IF NOT EXISTS idx_duels_created_at ON duels(created_at);
CREATE INDEX IF NOT EXISTS idx_duels_created_by ON duels(created_by_user_id);
CREATE INDEX IF NOT EXISTS idx_duel_participants_user ON duel_participants(user_id);
CREATE INDEX IF NOT EXISTS idx_game_rounds_duel ON game_rounds(duel_id);
CREATE INDEX IF NOT EXISTS idx_player_answers_round ON player_answers(round_id);
