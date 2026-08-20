# Hush frontend

React + Vite клиент для комнат и LiveKit.

## Запуск

```bash
cp .env.example .env
npm install
npm run dev
```

По умолчанию:

- REST API: `http://localhost:8080`
- Gateway WebSocket: `ws://localhost:8081/ws`
- Frontend: `http://localhost:5173`

## Ожидаемый ответ Gateway

```json
{
  "type": "joined",
  "data": {
    "room_id": "uuid",
    "livekit": {
      "url": "ws://localhost:7880",
      "token": "jwt"
    }
  }
}
```

## REST API

- `POST /user/init` — `{ "username": "..." }`
- `GET /rooms`
- `POST /rooms` — заголовок `X-User-ID`, тело `{ "name": "..." }`

`GET /rooms` может возвращать пользователей как `users`, `room_users` или `RoomUsers`; клиент нормализует вложенные `user_info`, `UserInfo` и `user`.
