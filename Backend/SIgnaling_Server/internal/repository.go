package internal

import (
	m "Signal_Server/Models"

	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type repoPart struct {
	rdb *redis.Client
}

func NewRepoPart(rdb *redis.Client) *repoPart {
	return &repoPart{rdb: rdb}
}

func (r repoPart) AddUser(ctx context.Context, roomId string, user *m.User) error {
	pipe := r.rdb.TxPipeline()

	pipe.HSet(ctx, "user:"+user.Id,
		"id:", user.Id,
		"user_name", user.UserName,
	)

	pipe.SAdd(ctx, "room:"+roomId+":users", user.Id)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("err:%w add user:%s into room:%s", err, user.Id, roomId)
	}

	return nil
}
