# LangDuel Cheat Sheet (1 страница)

## Главная идея
WS — это только транспорт. Логика игры — только в `duel`.

## Основные файлы
- `cmd/server/main.go` — запуск сервера.
- `internal/server/*` — HTTP маршруты.
- `internal/ws/*` — WebSocket транспорт.
- `internal/duel/*` — игровая логика.

## Поток данных
1. Клиент шлет `join` / `answer`.
2. `ws/handler.go` вызывает `duel.Manager`.
3. `duel.Manager` возвращает события.
4. `ws/hub.go` рассылает события только по комнате.

## Сообщения
Вход:
- `join`
- `answer`

Выход:
- `room_state`
- `round_start`
- `update`
- `round_end`
- `game_over`
- `player_joined`
- `player_left`
- `error`

## Где менять поведение
- Логику боя: `internal/duel/*`
- Протокол WS: `internal/ws/handler.go`
- Фразы: `internal/duel/room.go`
- UI: `battle.html`

## Быстрый запуск
```powershell
go run ./cmd/server
```
Открой `battle.html` в двух вкладках (u1/u2) и подключись.
