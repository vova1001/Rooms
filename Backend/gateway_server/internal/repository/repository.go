package repository

import (
	"database/sql"
	"time"

	"context"
	"fmt"

	m "backend/gateway_server/models"

	rm "backend/gateway_server/internal/repository/message"

	"github.com/redis/go-redis/v9"
)

type RepoPart struct {
	rdb     *redis.Client
	db      *sql.DB
	RepoMsg *rm.Repository
}

func NewRepoPart(rdb *redis.Client, db *sql.DB) *RepoPart {
	return &RepoPart{rdb: rdb, db: db, RepoMsg: rm.New(db)}
}

func (r RepoPart) AddUser(ctx context.Context, roomId string, user *m.User) error {
	pipe := r.rdb.TxPipeline()

	pipe.HSet(ctx, "user:"+user.Id,
		"id", user.Id,
		"user_name", user.UserName,
		"avatar", user.Avatar,
		"created_at", user.CreatedAt.Format(time.RFC3339Nano),
		"room_id", roomId,
	)

	pipe.SAdd(ctx, "room:"+roomId+":users", user.Id)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("err:%w add user:%s into room:%s", err, user.Id, roomId)
	}

	return nil
}

func (r RepoPart) DeleteUser(ctx context.Context, roomId, userId string) error {
	pipe := r.rdb.TxPipeline()

	pipe.SRem(ctx, "room:"+roomId+":users", userId)
	pipe.Del(ctx, "user:"+userId)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("err delete user:%s from room:%s err:%w", userId, roomId, err)
	}

	return nil
}

func (r RepoPart) GetSession(ctx context.Context, key string) ([]byte, error) {
	val, err := r.rdb.Get(ctx, key).Bytes()
	if err != nil {
		return nil, fmt.Errorf("get session from redis: %w", err)
	}

	return val, nil
}
