# Hush production deploy

Domains:
- https://hushh.site
- https://api.hushh.site
- wss://ws.hushh.site/ws
- wss://livekit.hushh.site

Required VPS firewall ports:
- 22/tcp SSH
- 80/tcp HTTP (Caddy/ACME)
- 443/tcp HTTPS/WSS
- 443/udp HTTP/3 (optional, enabled)
- 7881/tcp LiveKit WebRTC TCP fallback
- 7882/udp LiveKit WebRTC media

Do NOT expose PostgreSQL 5432 or Redis 6379 publicly.

Start:
```bash
cd /opt/hushh
docker compose config
docker compose up -d --build
docker compose ps
```

Logs:
```bash
docker compose logs --tail=100 caddy
docker compose logs --tail=100 api-server
docker compose logs --tail=100 gateway-server
docker compose logs --tail=100 livekit
```
