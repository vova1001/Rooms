package main

import (
	c "Signal_Server/Config"
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

func main() {
	cfg := c.LoadCfgDB()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAdr,
		Password: cfg.RedisPass,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis err ping%v", err)
	}

	// repo := i.NewRepoPart(rdb)

	fmt.Println("Signaling_UP!")
}
