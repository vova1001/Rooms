package repository

import (
	"context"
	"crypto/hmac"
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

	VerifyLimit int64 = 5
	VerifyKD          = 5 * time.Minute
)

var incrRateLimitScript = redis.NewScript(`
	local count = redis.call("INCR", KEYS[1])

	if count == 1 then 
		redis.call("PEXPIRE", KEYS[1], ARGV[1])
	end

	return count
`)

func (r *PartRepo) AllowVerifyByEmail(ctx context.Context, email string) (bool, error) {
	key := buildVerifyRateLimitKey(email, r.keySecret)

	count, err := incrRateLimitScript.Run(ctx, r.rdb, []string{key}, VerifyKD.Milliseconds()).Int64()
	if err != nil {
		return false, fmt.Errorf("execute Email rate limit script: %w", err)
	}

	return count <= VerifyLimit, nil
}

func (r *PartRepo) AllowByIp(ctx context.Context, rawIp string) (bool, error) {
	ip, err := normalizeIP(rawIp)
	if err != nil {
		return false, err
	}
	key := buildIPRateLimitKey(ip, r.keySecret)

	count, err := incrRateLimitScript.Run(ctx, r.rdb, []string{key}, IpKD.Milliseconds()).Int64()

	if err != nil {
		return false, fmt.Errorf("execute IP rate limit script: %w", err)
	}

	return count <= IpLimit, nil

}

func (r *PartRepo) AllowByEmail(ctx context.Context, email string) (bool, error) {
	key := buildEmailRateLimitKey(email, r.keySecret)

	count, err := incrRateLimitScript.Run(ctx, r.rdb, []string{key}, EmailKD.Milliseconds()).Int64()
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

func buildIPRateLimitKey(ip, secret string) string {
	return "rate_limit:auth:ip:" + hashKey(ip, secret)
}

func buildEmailRateLimitKey(email, secret string) string {
	return "rate_limit:auth:email:" + hashKey(email, secret)
}

func buildVerifyRateLimitKey(email, secret string) string {
	return "rate_limit:auth:verify:" + hashKey(email, secret)
}

func hashKey(key, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}
