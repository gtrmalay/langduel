# AI Генерация фраз

## Концепция

При создании комнаты игрок выбирает тему и сложность, затем нажимает "Сгенерировать". AI создаёт пачку фраз с вариациями правильных ответов.

## Направления перевода

1. **EN → RU** (английский → русский)
2. **RU → EN** (русский → английский)

## Темы

| Тема | Описание |
|------|---------|
| default | Общие фразы |
| movies | Кино и сериалы |
| travel | Путешествия |
| food | Еда и рестораны |
| sports | Спорт |
| slang | Сленг (advanced) |

## Сложность

| Уровень | Описание |
|---------|---------|
| beginner | Простые фразы |
| intermediate | Средние фразы |
| advanced | Сложные + сленг |

## Процесс генерации

```
Игрок выбирает тему/сложность
           ↓
    [⚡ СГЕНЕРИРОВАТЬ]
           ↓
    AI генерирует 20-30 фраз
    (~5-10 секунд)
           ↓
    Фразы сохраняются в БД
           ↓
    Дуэль начинается
```

## Промпт для AI

```
Сгенерируй 25 фраз для языковой дуэли.

Тема: {topic}
Сложность: {difficulty}
Направление: {lang_from} → {lang_to}

Формат JSON (только массив):
[
  {
    "prompt": "фраза на исходном языке",
    "answers": ["вариант1", "вариант2", "вариант3", "вариант4"]
  }
]

Требования:
- Фразы реальные и частые в речи
- Для advanced включай сленг и идиомы
- 3-5 вариаций правильного перевода
- Вариации грамматически верные
```

## Структура в БД

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

## Валидация ответов

```go
func CheckAnswer(userAnswer string, correctAnswers []string) bool {
    // 1. Точное совпадение (case-insensitive, trim)
    for _, answer := range correctAnswers {
        if strings.EqualFold(strings.TrimSpace(userAnswer), answer) {
            return true
        }
    }
    
    // 2. Fuzzy matching (допуск опечаток)
    for _, answer := range correctAnswers {
        if levenshteinDistance(userAnswer, answer) <= 2 {
            return true
        }
    }
    
    return false
}
```

## AI Провайдеры

| Провайдер | Стоимость | Качество | Примечание |
|-----------|-----------|---------|-----------|
| Grok (xAI) | $5/месяц | Хорошее | Рекомендуется |
| OpenAI GPT-3.5 | $5/месяц | Отличное | Стандарт |
| Claude | $20/месяц | Лучшее | Избыточно |
| Ollama (локально) | Бесплатно | Среднее | Для тестов |

## Обработка ошибок

- Таймаут генерации: 30 секунд
- При ошибке AI: fallback на статические фразы
- Rate limiting: не более 10 генераций в минуту
