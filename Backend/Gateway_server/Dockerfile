# ---- build stage ----
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /out/gateway-server ./main.go

# ---- runtime stage ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/gateway-server ./gateway-server
COPY .env.docker ./.env

EXPOSE 8081

CMD ["./gateway-server"]
