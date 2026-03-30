# AI Strategy - Генерация и проверка фраз

## Подход

Генерация фраз происходит **при создании комнаты** (pre-generation). AI создаёт пачку фраз с вариациями ответов, которые сохраняются в БД.

```
При создании комнаты:
┌─────────────────────────────────────────┐
│  Тема: movies, Сложность: intermediate  │
│                      [⚡ СГЕНЕРИРОВАТЬ] │
└─────────────────────────────────────────┘
                        ↓
              AI генерирует 20-30 фраз
              с вариациями правильных ответов
                        ↓
              Сохраняются в БД (мгновенно)
                        ↓
              Дуэль начинается
```

## Преимущества

| | |
|--|--|
| ✅ Скорость | Нет задержки во время игры |
| ✅ Надёжность | AI API может упасть - фразы уже в БД |
| ✅ Контроль | Можно проверить фразы перед игрой |
| ✅ Вариации | Несколько правильных ответов |

## Структура БД

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

## Проверка ответов

```go
func CheckAnswer(userAnswer string, correctAnswers []string) bool {
    // 1. Точное совпадение (case-insensitive, trim)
    for _, answer := range correctAnswers {
        if strings.EqualFold(strings.TrimSpace(userAnswer), answer) {
            return true
        }
    }
    
    // 2. Fuzzy matching (Levenshtein distance ≤ 2)
    for _, answer := range correctAnswers {
        if levenshteinDistance(userAnswer, answer) <= 2 {
            return true
        }
    }
    
    return false
}
```

## AI Провайдеры

| Провайдер | Стоимость | Примечание |
|-----------|-----------|-----------|
| Grok (xAI) | $5/месяц | Рекомендуется |
| OpenAI GPT-3.5 | $5/месяц | Стандарт |
| Ollama | Бесплатно | Для тестов |

## Промпт для генерации

```
Сгенерируй 25 фраз для языковой дуэли.

Тема: {topic}
Сложность: {difficulty}
Направление: {lang_from} → {lang_to}

Формат JSON (только массив):
[{"prompt": "...", "answers": ["...", "...", "..."]}]

Требования:
- Фразы реальные и частые в речи
- Для advanced включай сленг и идиомы
- 3-5 вариаций правильного перевода
```

## Рекомендации

1. **Grok или GPT-3.5** - достаточное качество за разумную цену
2. **Fallback** - при ошибке AI использовать статические фразы
3. **Таймаут** - 30 секунд максимум на генерацию
