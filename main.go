package main

import (
	"fmt"
	"log"
	"net/http"
	d "rooms/Backend/Api_Server/db"
	internal "rooms/Backend/Api_Server/internal"
	c "rooms/Backend/Api_Server/internal/config"
	"time"
)

func main() {
	cfg, err := c.LoadCfgDB()
	if err != nil {
		log.Fatalf("err load cfg:%v", err)
	}

	db, err := d.DBinit(cfg)
	if err != nil {
		log.Fatalf("err conect from db: %v", err)
	}
	if err := d.Migrate(db); err != nil {
		log.Fatalf("migrate err:%v", err)
	}

	repo := internal.NewRepo(db)
	service := internal.NewService(repo)
	handler := internal.NewHandler(service)

	mux := http.DefaultServeMux

	handler.RegisterRote(mux)

	server := http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	fmt.Println("Api server is up!")
	log.Fatal(server.ListenAndServe(), "Server is dead")
}
