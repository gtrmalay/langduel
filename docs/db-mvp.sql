-- LangDuel MVP schema (PostgreSQL)
-- Focus: users, duels, participants, rounds, answers, phrases.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(30) NOT NULL UNIQUE,
  email VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255),
  is_guest BOOLEAN NOT NULL DEFAULT FALSE,
  guest_expires_at TIMESTAMPTZ,
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
