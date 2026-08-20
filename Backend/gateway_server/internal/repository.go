package internal

import (
	"database/sql"
	"time"

	"context"
	"fmt"

	m "backend/gateway_server/models"

	"github.com/redis/go-redis/v9"
)

type repoPart struct {
	rdb *redis.Client
	db  *sql.DB
}

func NewRepoPart(rdb *redis.Client, db *sql.DB) *repoPart {
	return &repoPart{rdb: rdb, db: db}
}

func (r repoPart) AddUser(ctx context.Context, roomId string, user *m.User) error {
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

func (r repoPart) DeleteUser(ctx context.Context, roomId, userId string) error {
	pipe := r.rdb.TxPipeline()

	pipe.SRem(ctx, "room:"+roomId+":users", userId)
	pipe.Del(ctx, "user:"+userId)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("err delete user:%s from room:%s err:%w", userId, roomId, err)
	}

	return nil
}
