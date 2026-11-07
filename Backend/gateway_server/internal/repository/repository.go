package repository

import (
	"database/sql"
	"errors"
	"time"

	"context"
	"fmt"

	m "backend/gateway_server/models"

	rm "backend/gateway_server/internal/repository/message"
	rl "backend/gateway_server/internal/repository/ratelimit"

	"github.com/redis/go-redis/v9"
)

type RepoPart struct {
	rdb       *redis.Client
	db        *sql.DB
	RepoMsg   *rm.Repository
	RateLimit *rl.Repository
}

func NewRepoPart(rdb *redis.Client, db *sql.DB) *RepoPart {
	return &RepoPart{rdb: rdb, db: db, RepoMsg: rm.New(db), RateLimit: rl.New(rdb)}
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

func (r RepoPart) CheckRoomAccess(ctx context.Context, roomID, userID string) (bool, bool, error) {
	var allowed bool
	err := r.db.QueryRowContext(ctx, `
		SELECT NOT r.is_private
		       OR r.owner_id = $2
		       OR EXISTS (SELECT 1 FROM room_users ru WHERE ru.room_id = r.id AND ru.user_id = $2)
		FROM rooms r
		WHERE r.id = $1`, roomID, userID,
	).Scan(&allowed)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, false, nil
		}
		return false, false, fmt.Errorf("check room access: %w", err)
	}
	return true, allowed, nil
}
