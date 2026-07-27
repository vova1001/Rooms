package main

import (
	c "GateWay/config"
	i "GateWay/internal"
	"GateWay/livekit"
	"context"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

func main() {

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(".env not found")
	}

	cfg := c.LoadCfgDB()
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAdr,
		Password: cfg.RedisPass,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis err ping%v", err)
	}

	cfgLK := c.LoadCfgLK()
	LKS := livekit.NewTokenService(cfgLK.LKurl, cfgLK.LKapiKey, cfgLK.LKapiSecret)

	log.Printf(
		"redis addr=%q passEmpty=%v",
		cfg.RedisAdr,
		cfg.RedisPass == "",
	)

	log.Printf(
		"livekit url=%q key=%q secretEmpty=%v",
		cfgLK.LKurl,
		cfgLK.LKapiKey,
		cfgLK.LKapiSecret == "",
	)

	repo := i.NewRepoPart(rdb)
	service := i.NewService(repo, LKS)
	handler := i.NewHandler(service)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", handler.ServeHTTP)

	server := http.Server{
		Addr:    ":8081",
		Handler: mux,
	}

	log.Println("Geteway starting!!!")
	log.Fatal(server.ListenAndServe())

}
