package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	config "backend/api_server/config"
	cors "backend/api_server/cors"

	han "backend/api_server/internal/handler"
	repo "backend/api_server/internal/repository"
	serv "backend/api_server/internal/service"
	e "backend/api_server/internal/service/email"
	"backend/api_server/internal/service/otp"
	s "backend/api_server/internal/service/session"

	configDB "backend/shared/configDB"
	d "backend/shared/postgre"

	"github.com/redis/go-redis/v9"
)

func main() {
	config.LoadEnv()

	cfg, err := configDB.LoadCfgDB()
	if err != nil {
		log.Fatalf("err load DB config: %v", err)
	}

	cfgRDB, err := configDB.LoadCfgRDB()
	if err != nil {
		log.Fatalf("err load Redis config: %v", err)
	}

	cfgSMTP, err := config.LoadCfgSMTP()
	if err != nil {
		log.Fatalf("err load SMTP config: %v", err)
	}

	cfgOTP, err := config.LoadCfgOTP()
	if err != nil {
		log.Fatalf("err load OTP config: %v", err)
	}

	dbConn, err := d.DBinit(cfg)
	if err != nil {
		log.Fatalf("err connect to db: %v", err)
	}

	if err := d.Migrate(dbConn); err != nil {
		log.Fatalf("migrate err: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     cfgRDB.RedisAdr,
		Password: cfgRDB.RedisPass,
	})

	repository := repo.NewRepo(dbConn, rdb)

	generator := otp.NewGenerator()
	hasher := otp.NewCodeHasher(cfgOTP.Secret)
	otpService := otp.NewService(generator, hasher)

	generatorSession := s.NewGeneratorS()
	hasherSession := s.NewHasher()
	serviceSession := s.NewService(repository, generatorSession, hasherSession)

	sender := e.NewSTMP(
		cfgSMTP.SMTPHost,
		cfgSMTP.SMTPFrom,
		cfgSMTP.SMTPPort,
		cfgSMTP.SMTPUsername,
		cfgSMTP.SMTPPassword,
	)

	service := serv.NewService(
		repository,
		otpService,
		repository,
		sender,
		serviceSession,
	)

	handler := han.NewHandler(service)

	router := http.NewServeMux()
	handler.RegisterRoutes(router)

	fileServer := http.FileServer(http.Dir("./static"))

	router.Handle("/static/", http.StripPrefix("/static/", fileServer))

	server := http.Server{
		Addr:         ":8080",
		Handler:      cors.CORS(router),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	fmt.Println("API server is up on :8080")
	log.Fatal(server.ListenAndServe())
}
