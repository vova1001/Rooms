package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"sync"
	"testing"

	han "backend/api_server/internal/handler"
	repo "backend/api_server/internal/repository"
	serv "backend/api_server/internal/service"
	"backend/api_server/internal/service/otp"
	s "backend/api_server/internal/service/session"
	sharedGetAuthS "backend/shared/getAuthSession"
	d "backend/shared/postgre"

	"github.com/redis/go-redis/v9"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

// Фейковый SMTP: не шлёт письмо, а запоминает код
type captureSender struct {
	mu    sync.Mutex
	codes map[string]string
}

func (c *captureSender) SendCode(_ context.Context, email, code string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.codes[email] = code
	return nil
}

func (c *captureSender) last(email string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.codes[email]
}

// Поднимает Postgres + Redis и собирает сервер как в main.go
func newTestServer(t *testing.T) (*httptest.Server, *captureSender) {
	t.Helper()
	ctx := context.Background()

	pg, err := tcpostgres.Run(ctx, "postgres:16-alpine",
		tcpostgres.WithDatabase("test"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pg.Terminate(ctx) })

	dsn, _ := pg.ConnectionString(ctx, "sslmode=disable")
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatal(err)
	}
	if err := d.Migrate(db); err != nil {
		t.Fatal(err)
	}

	rc, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rc.Terminate(ctx) })

	uri, _ := rc.ConnectionString(ctx)
	opts, _ := redis.ParseURL(uri)
	rdb := redis.NewClient(opts)

	// --- та же сборка, что в main.go ---
	repository := repo.NewRepo(db, rdb, "test-secret")
	otpService := otp.NewService(otp.NewGenerator(), otp.NewCodeHasher("test-secret"))
	hasherSession := sharedGetAuthS.NewHasher()
	serviceSession := s.NewService(repository, s.NewGeneratorS(), hasherSession)
	sender := &captureSender{codes: map[string]string{}}

	usersRepo := sharedRepoUsers.NewRepository(db)

	service := serv.NewService(
		repository,     // repo
		otpService,     // otp
		repository,     // otpRepo
		sender,         // sender
		serviceSession, // sessions
		repository,     // limiter
		usersRepo,
	)
	handler := han.NewHandler(service,
		sharedGetAuthS.NewReader(repository, hasherSession),
		usersRepo,
	)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, sender
}

// Клиент с «браузерными» cookie
func newClient() *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar}
}

func postJSON(t *testing.T, c *http.Client, url string, body any) *http.Response {
	t.Helper()
	b, _ := json.Marshal(body)
	resp, err := c.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

// ---------- Сценарий 1: регистрация от начала до конца ----------

func TestAuthFlow_NewUserRegisters(t *testing.T) {
	srv, sender := newTestServer(t)
	c := newClient()
	email := "new@user.com"

	resp := postJSON(t, c, srv.URL+"/auth/email/send-code", map[string]string{"email": email})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("send-code: %d", resp.StatusCode)
	}

	code := sender.last(email)
	resp = postJSON(t, c, srv.URL+"/auth/email/verify-code", map[string]string{"email": email, "code": code})
	var verify struct {
		RequiresRegister bool `json:"requires_register"`
	}
	json.NewDecoder(resp.Body).Decode(&verify)
	if !verify.RequiresRegister {
		t.Fatal("новый пользователь должен попасть на регистрацию")
	}

	resp = postJSON(t, c, srv.URL+"/create-user", map[string]string{"username": "vova"})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create-user: %d", resp.StatusCode)
	}

	resp, _ = c.Get(srv.URL + "/auth/session")
	var session struct {
		Status string `json:"status"`
	}
	json.NewDecoder(resp.Body).Decode(&session)
	if session.Status != "authorized" {
		t.Fatalf("ожидали authorized, получили %q", session.Status)
	}
}

// ---------- Сценарий 2: перебор кода блокируется ----------

func TestAuthFlow_BruteForceBlocked(t *testing.T) {
	srv, sender := newTestServer(t)
	c := newClient()
	email := "victim@user.com"

	postJSON(t, c, srv.URL+"/auth/email/send-code", map[string]string{"email": email})
	realCode := sender.last(email)

	for i := 0; i < 5; i++ {
		postJSON(t, c, srv.URL+"/auth/email/verify-code", map[string]string{"email": email, "code": "000000"})
	}

	resp := postJSON(t, c, srv.URL+"/auth/email/verify-code", map[string]string{"email": email, "code": realCode})
	if resp.StatusCode == http.StatusOK {
		t.Fatal("после 5 попыток вход должен быть заблокирован")
	}
}
