# syntax=docker/dockerfile:1

FROM node:22-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./

RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund

COPY . .

ARG VITE_API_URL
ARG VITE_GATEWAY_WS

ENV VITE_API_URL=$VITE_API_URL
ENV VITE_GATEWAY_WS=$VITE_GATEWAY_WS

RUN npm run build


FROM nginx:1.27-alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]