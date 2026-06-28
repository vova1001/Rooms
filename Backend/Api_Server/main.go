package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"rooms/config"
	"rooms/db"
	"rooms/internal"

	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.LoadCfgDB()
	if err != nil {
		log.Fatalf("err load cfg: %v", err)
	}

	dbConn, err := db.DBinit(cfg)
	if err != nil {
		log.Fatalf("err connect from db: %v", err)
	}
	if err := db.Migrate(dbConn); err != nil {
		log.Fatalf("migrate err: %v", err)
	}

	repo := internal.NewRepo(dbConn)
	service := internal.NewService(repo)
	handler := internal.NewHandler(service)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	server := http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Println("API server is up on :8080")
	log.Fatal(server.ListenAndServe())
}
