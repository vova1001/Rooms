// e2e_test.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"rooms/db"
	"rooms/internal"
	"rooms/model"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestE2E_Rooms(t *testing.T) {
	// Запускаем PostgreSQL в Docker
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "14",
		Env: []string{
			"POSTGRES_USER=test",
			"POSTGRES_PASSWORD=test",
			"POSTGRES_DB=testdb",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)

	// Ждём, пока БД поднимется
	var dbURL string
	if err := pool.Retry(func() error {
		dbURL = fmt.Sprintf("postgres://test:test@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))
		return db.InitTestDB(dbURL) // напишем эту функцию ниже
	}); err != nil {
		t.Fatalf("Could not connect to database: %s", err)
	}
	defer func() {
		if err := pool.Purge(resource); err != nil {
			t.Logf("Could not purge resource: %s", err)
		}
	}()

	// Инициализируем тестовую БД (создаём таблицы)
	testDB := db.GetDB() // предположим, db.InitTestDB сохранила глобальное подключение
	defer testDB.Close()

	// Создаём репозиторий, сервис, хендлер и сервер
	repo := internal.NewRepo(testDB)
	service := internal.NewService(repo)
	handler := internal.NewHandler(service)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("Server error: %v", err)
		}
	}()
	defer srv.Close()
	time.Sleep(1 * time.Second) // даём серверу время запуститься

	baseURL := "http://localhost:8081"

	// --- Тест 1: инициализация пользователя ---
	resp, err := http.Post(baseURL+"/user/init", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var userResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&userResp)
	require.NoError(t, err)
	userIDStr, ok := userResp["id"].(string)
	require.True(t, ok)
	userID, err := uuid.Parse(userIDStr)
	require.NoError(t, err)

	// --- Тест 2: создание комнаты ---
	roomReq := model.CreateRoomRequest{Name: "Test Room"}
	body, _ := json.Marshal(roomReq)
	req, _ := http.NewRequest("POST", baseURL+"/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var roomResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&roomResp)
	require.NoError(t, err)
	roomIDStr := roomResp["id"].(string)
	roomID, err := uuid.Parse(roomIDStr)
	require.NoError(t, err)
	assert.Equal(t, "Test Room", roomResp["name"])
	assert.Equal(t, userID.String(), roomResp["owner_id"])

	// --- Тест 3: получение всех комнат ---
	resp, err = http.Get(baseURL + "/rooms")
	require.NoError(t, err)
	defer resp.Body.Close()
	var rooms []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&rooms)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rooms), 1)

	// --- Тест 4: получение пользователей комнаты ---
	resp, err = http.Get(baseURL + "/rooms/" + roomID.String() + "/users")
	require.NoError(t, err)
	defer resp.Body.Close()
	var users []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(t, err)
	assert.Len(t, users, 1) // только владелец
	assert.Equal(t, userID.String(), users[0]["id"])

	// --- Тест 5: негативный сценарий – создание комнаты без имени ---
	badReq := model.CreateRoomRequest{Name: ""}
	body, _ = json.Marshal(badReq)
	req, _ = http.NewRequest("POST", baseURL+"/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID.String())
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// --- Тест 6: создание комнаты с несуществующим ownerID ---
	badOwner := uuid.New()
	req, _ = http.NewRequest("POST", baseURL+"/rooms", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", badOwner.String())
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode) // foreign key
}
