# ---- build stage ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

# кэшируем зависимости отдельным слоем
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./main.go

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/api-server ./api-server
# .env нужен, т.к. код падает с Fatal, если файла нет (даже если
# переменные уже заданы через docker-compose environment/env_file)
COPY .env.docker ./.env

EXPOSE 8080

CMD ["./api-server"]
