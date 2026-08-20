package config

import "os"

type ConfigLK struct {
	LKurl       string
	LKapiKey    string
	LKapiSecret string
}

func LoadCfgLK() *ConfigLK {
	return &ConfigLK{
		LKurl:       os.Getenv("LIVEKIT_URL"),
		LKapiKey:    os.Getenv("LIVEKIT_API_KEY"),
		LKapiSecret: os.Getenv("LIVEKIT_API_SECRET"),
	}
}
