//go:build integration

package ratelimit

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func newTestRepo(t *testing.T) (*Repository, *redis.Client) {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		t.Skip("REDIS_ADDR не задан")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: os.Getenv("REDIS_PASS"),
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("redis недоступен: %v", err)
	}
	t.Cleanup(func() { rdb.Close() })

	return New(rdb), rdb
}

func testKey(t *testing.T, rdb *redis.Client) string {
	key := "rate_limit:test:" + uuid.NewString()
	t.Cleanup(func() { rdb.Del(context.Background(), key) })
	return key
}

func TestAllow_TenPassEleventhBlocked(t *testing.T) {
	repo, rdb := newTestRepo(t)
	key := testKey(t, rdb)
	ctx := context.Background()

	for i := 1; i <= 10; i++ {
		ok, err := repo.Allow(ctx, key, 10, 5*time.Second)
		if err != nil {
			t.Fatalf("сообщение %d: %v", i, err)
		}
		if !ok {
			t.Fatalf("сообщение %d должно пройти", i)
		}
	}

	ok, err := repo.Allow(ctx, key, 10, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("11-е сообщение должно быть заблокировано")
	}
}

func TestAllow_UsersAreIndependent(t *testing.T) {
	repo, rdb := newTestRepo(t)
	vasya := testKey(t, rdb)
	petya := testKey(t, rdb)
	ctx := context.Background()

	for i := 0; i < 10; i++ {
		repo.Allow(ctx, vasya, 10, 5*time.Second)
	}

	ok, err := repo.Allow(ctx, petya, 10, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("лимит Васи не должен блокировать Петю")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	repo, rdb := newTestRepo(t)
	key := testKey(t, rdb)
	ctx := context.Background()
	window := 500 * time.Millisecond

	for i := 0; i < 11; i++ {
		repo.Allow(ctx, key, 10, window)
	}

	time.Sleep(window + 100*time.Millisecond)

	ok, err := repo.Allow(ctx, key, 10, window)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("после окна лимит должен сброситься")
	}
}

func TestAllow_KeyHasTTL(t *testing.T) {
	repo, rdb := newTestRepo(t)
	key := testKey(t, rdb)
	ctx := context.Background()

	repo.Allow(ctx, key, 10, 5*time.Second)

	ttl, err := rdb.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= 0 || ttl > 5*time.Second {
		t.Fatalf("у ключа должен быть срок жизни до 5с, а он %v", ttl)
	}
}
