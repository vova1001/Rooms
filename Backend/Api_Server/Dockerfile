FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/api-server ./main.go


FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/api-server ./api-server

# Копируем аватарки
COPY --from=builder /app/static ./static

COPY .env.docker ./.env

EXPOSE 8080

CMD ["./api-server"]