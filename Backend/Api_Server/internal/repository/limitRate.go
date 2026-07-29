package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	IpLimit int64         = 10
	IpKD    time.Duration = 15 * time.Minute

	EmailLimit int64 = 3
	EmailKD          = 15 * time.Minute
)

var incrRateLimitScript = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])

	if count == 1 then 
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end

	return count
`)

func (r *PartRepo) AllowByIp(ctx context.Context, rawIp string) (bool, error) {
	ip, err := normalizeIP(rawIp)
	if err != nil {
		return false, err
	}
	key := buildIPRateLimitKey(ip)

	count, err := incrRateLimitScript.Run(ctx, r.rdb, []string{key}, IpKD).Int64()

	if err != nil {
		return false, fmt.Errorf("execute IP rate limit script: %w", err)
	}

	return count <= IpLimit, nil

}

func (r *PartRepo) AllowByEmail(ctx context.Context, email string) (bool, error) {
	key := buildEmailRateLimitKey(email)

	count, err := incrRateLimitScript.Run(ctx, r.rdb, []string{key}, EmailKD).Int64()
	if err != nil {
		return false, fmt.Errorf("execute Email rate limit script: %w", err)
	}

	return count <= EmailLimit, nil
}

func normalizeIP(rawIP string) (string, error) {
	ip := net.ParseIP(rawIP)
	if ip == nil {
		return "", fmt.Errorf("Invalid Ip")
	}
	return ip.String(), nil
}

func buildIPRateLimitKey(ip string) string {
	return "rate_limit:auth:ip:" + hashIP(ip)
}

func buildEmailRateLimitKey(email string) string {
	return "rate_limit:auth:email:" + hashIP(email)
}

func hashIP(ip string) string {
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}
