# LangDuel API (MVP)

## WebSocket endpoint
`ws://localhost:8080/ws`

## Входящие сообщения (клиент -> сервер)

### join
```json
{"type":"join","room_id":"room1","user_id":"u1","lang":"en","topic":"default"}
```

### answer
```json
{"type":"answer","room_id":"room1","user_id":"u1","answer":"kot","speed":1200}
```

## Исходящие события (сервер -> клиент)

### room_state
```json
{"type":"room_state","room_id":"room1","round":1,"round_token":1,"prompt":"cat","players":["u1","u2"],"hp":{"u1":100,"u2":100}}
```

### player_joined
```json
{"type":"player_joined","room_id":"room1","players":["u1","u2"],"hp":{"u1":100,"u2":100}}
```

### player_left
```json
{"type":"player_left","room_id":"room1","players":["u1"],"hp":{"u1":100},"reason":"disconnect"}
```

### round_start
```json
{"type":"round_start","room_id":"room1","round":1,"round_token":1,"prompt":"cat","hp":{"u1":100,"u2":100}}
```

### round_end (timeout)
```json
{"type":"round_end","room_id":"room1","round":1,"round_token":1,"prompt":"cat","reason":"timeout","hp":{"u1":100,"u2":100}}
```

### update
```json
{"type":"update","room_id":"room1","attacker_id":"u1","defender_id":"u2","damage":15,"correct":true,"hp":{"u1":100,"u2":85}}
```

### game_over
```json
{"type":"game_over","room_id":"room1","winner_id":"u1","hp":{"u1":100,"u2":0}}
```

### error
```json
{"type":"error","room_id":"room1","error":"join room first"}
```

## Правила MVP
- Комната максимум 2 игрока.
- Раунд стартует при 2 игроках.
- Таймер раунда 10 секунд.
- При таймауте урон не наносится, просто начинается новый раунд.
