# ---- build stage ----
FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .

# Адреса API/Gateway "зашиваются" в статическую сборку Vite на этапе билда.
# Значения по умолчанию рассчитаны на локальный docker-compose запуск,
# при необходимости переопредели через --build-arg или переменные в docker-compose.yml
ARG VITE_API_URL=http://localhost:8080
ARG VITE_GATEWAY_WS=ws://localhost:8081/ws
ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GATEWAY_WS=$VITE_GATEWAY_WS

RUN npm run build

# ---- runtime stage ----
FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
