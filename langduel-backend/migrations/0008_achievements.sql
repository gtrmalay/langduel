-- achievements: достижения игроков
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

-- Базовые достижения
INSERT INTO achievements (id, name, description, icon, xp_reward) VALUES
('first_win', 'Первая победа', 'Выиграйте первый матч', '🏆', 10),
('warrior', 'Воин', 'Выиграйте 10 матчей', '⚔️', 25),
('veteran', 'Ветеран', 'Выиграйте 50 матчей', '🛡️', 50),
('champion', 'Чемпион', 'Выиграйте 100 матчей', '👑', 100),
('streak_5', 'Натиск', '5 побед подряд', '🔥', 20),
('streak_10', 'Мастер натиска', '10 побед подряд', '💥', 50),
('games_10', 'Новичок', 'Сыграйте 10 матчей', '🎮', 10),
('games_50', 'Игрок', 'Сыграйте 50 матчей', '🎯', 25);

CREATE INDEX idx_user_achievements_user ON user_achievements(user_id);
