# Database Schema

## Существующие таблицы

### users
```sql
CREATE TABLE users (
  user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  username VARCHAR(30) NOT NULL UNIQUE,
  email VARCHAR(100) UNIQUE,
  password_hash VARCHAR(255),
  is_guest BOOLEAN DEFAULT FALSE,
  avatar VARCHAR(50) DEFAULT 'default',
  guest_expires_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### duels
```sql
CREATE TABLE duels (
  duel_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  room_code VARCHAR(32) UNIQUE NOT NULL,
  created_by_user_id UUID REFERENCES users(user_id),
  winner_user_id UUID REFERENCES users(user_id),
  language_from VARCHAR(10) DEFAULT 'en',
  language_to VARCHAR(10) DEFAULT 'ru',
  theme VARCHAR(30) DEFAULT 'default',
  difficulty INT DEFAULT 2,
  max_rounds INT DEFAULT 10,
  status VARCHAR(20) DEFAULT 'waiting',
  created_at TIMESTAMPTZ DEFAULT NOW(),
  started_at TIMESTAMPTZ,
  finished_at TIMESTAMPTZ
);
```

### duel_participants
```sql
CREATE TABLE duel_participants (
  participant_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  duel_id UUID REFERENCES duels(duel_id) ON DELETE CASCADE,
  user_id UUID REFERENCES users(user_id),
  player_order INT NOT NULL,
  initial_hp INT DEFAULT 100,
  final_hp INT,
  score INT DEFAULT 0,
  joined_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (duel_id, user_id)
);
```

### game_rounds
```sql
CREATE TABLE game_rounds (
  round_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  duel_id UUID REFERENCES duels(duel_id) ON DELETE CASCADE,
  round_number INT NOT NULL,
  phrase_id UUID REFERENCES phrases(phrase_id),
  time_limit_ms INT DEFAULT 10000,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  UNIQUE (duel_id, round_number)
);
```

### player_answers
```sql
CREATE TABLE player_answers (
  answer_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  round_id UUID REFERENCES game_rounds(round_id) ON DELETE CASCADE,
  participant_id UUID REFERENCES duel_participants(participant_id) ON DELETE CASCADE,
  translation_text TEXT NOT NULL,
  is_correct BOOLEAN DEFAULT FALSE,
  response_time_ms INT NOT NULL,
  damage_dealt INT DEFAULT 0,
  submitted_at TIMESTAMPTZ DEFAULT NOW()
);
```

### user_statistics
```sql
CREATE TABLE user_statistics (
  stats_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID UNIQUE REFERENCES users(user_id) ON DELETE CASCADE,
  total_duels_played INT DEFAULT 0,
  total_duels_won INT DEFAULT 0,
  total_score BIGINT DEFAULT 0,
  overall_accuracy FLOAT DEFAULT 0,
  best_win_streak INT DEFAULT 0,
  total_play_time_min INT DEFAULT 0,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Новые таблицы

### user_ratings (Этап 1)
```sql
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
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### ai_phrases (Этап 2)
```sql
CREATE TABLE ai_phrases (
  phrase_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  duel_id UUID REFERENCES duels(duel_id),
  prompt VARCHAR(255) NOT NULL,
  answers TEXT[] NOT NULL,
  topic VARCHAR(30) NOT NULL,
  difficulty INT NOT NULL DEFAULT 1,
  lang_from VARCHAR(10) NOT NULL,
  lang_to VARCHAR(10) NOT NULL,
  used BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### achievements (Этап 3)
```sql
CREATE TABLE achievements (
  id VARCHAR(50) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  icon VARCHAR(50),
  xp_reward INT DEFAULT 0
);

CREATE TABLE user_achievements (
  user_id UUID REFERENCES users(user_id) ON DELETE CASCADE,
  achievement_id VARCHAR(50) REFERENCES achievements(id),
  unlocked_at TIMESTAMPTZ DEFAULT NOW(),
  PRIMARY KEY (user_id, achievement_id)
);
```

### user_progress (Этап 4)
```sql
CREATE TABLE user_progress (
  user_id UUID PRIMARY KEY REFERENCES users(user_id) ON DELETE CASCADE,
  xp INT DEFAULT 0,
  level INT DEFAULT 1,
  updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Ранги (ELO)

| Звание | ELO | Цвет |
|--------|-----|------|
| 🥉 Newbie | 0-999 | #CD7F32 |
| 🥈 Apprentice | 1000-1999 | #C0C0C0 |
| 🥇 Expert | 2000-2999 | #FFD700 |
| 💎 Master | 3000+ | #B9F2FF |
| 😔 Struggler | 0 | special (если >10 поражений при 0) |

### Формула ELO
- Победа: +25 ELO
- Поражение: -15 ELO
- Минимум: 0

---

## Миграции

| # | Файл | Описание |
|---|------|---------|
| 0001 | init.sql | Базовая схема |
| 0002 | seed_phrases.sql | Начальные фразы |
| 0003 | room_code_length.sql | Длина room_code |
| 0004 | user_statistics.sql | Таблица статистики |
| 0005 | core_finalize.sql | Финализация |
| 0006 | user_avatar.sql | Поле avatar |
| 0007 | user_ratings.sql | ELO и ранги |
| 0008 | ai_phrases.sql | AI сгенерированные фразы |
| 0009 | achievements.sql | Достижения |
| 0010 | user_progress.sql | XP и уровень |
