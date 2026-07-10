package config

import "os"

type ConfigDB struct {
	RedisAdr  string
	RedisPass string
}

func LoadCfgDB() *ConfigDB {
	return &ConfigDB{
		RedisAdr:  os.Getenv("REDIS_ADDR"),
		RedisPass: os.Getenv("REDIS_PASS"),
	}
}
