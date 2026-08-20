package main

import (
	c "backend/gateway_server/config"
	i "backend/gateway_server/internal"
	"backend/gateway_server/livekit"
	configDB "backend/shared/configDB"
	d "backend/shared/postgre"

	"context"
	"log"
	"net/http"

	"github.com/redis/go-redis/v9"
)

func main() {

	cfgRDB, err := configDB.LoadCfgRDB()
	if err != nil {
		log.Printf("err cfgRDB: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfgRDB.RedisAdr,
		Password: cfgRDB.RedisPass,
	})
	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis err ping%v", err)
	}

	cfgDB, err := configDB.LoadCfgDB()
	if err != nil {
		log.Printf("err cfgDB: %v", err)
	}

	db, err := d.DBinit(cfgDB)
	if err != nil {
		log.Fatalf("err init db: %v", err)
	}

	cfgLK := c.LoadCfgLK()
	LKS := livekit.NewTokenService(cfgLK.LKurl, cfgLK.LKapiKey, cfgLK.LKapiSecret)

	log.Printf(
		"redis addr=%q passEmpty=%v",
		cfgRDB.RedisAdr,
		cfgRDB.RedisPass == "",
	)

	log.Printf(
		"livekit url=%q key=%q secretEmpty=%v",
		cfgLK.LKurl,
		cfgLK.LKapiKey,
		cfgLK.LKapiSecret == "",
	)

	repo := i.NewRepoPart(rdb, db)
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
