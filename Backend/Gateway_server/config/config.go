package config

import "os"

type ConfigDB struct {
	RedisAdr  string
	RedisPass string
}

type ConfigLK struct {
	LKurl       string
	LKapiKey    string
	LKapiSecret string
}

func LoadCfgDB() *ConfigDB {
	return &ConfigDB{
		RedisAdr:  os.Getenv("REDIS_ADDR"),
		RedisPass: os.Getenv("REDIS_PASS"),
	}
}

func LoadCfgLK() *ConfigLK {
	return &ConfigLK{
		LKurl:       os.Getenv("LIVEKIT_URL"),
		LKapiKey:    os.Getenv("LIVEKIT_API_KEY"),
		LKapiSecret: os.Getenv("LIVEKIT_API_SECRET"),
	}
}
