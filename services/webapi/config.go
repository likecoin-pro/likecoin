package webapi

import (
	"flag"
)

type Config struct {
	HTTPConnStr string
}

func NewConfig() *Config {
	cfg := &Config{
		HTTPConnStr: ":8666",
	}
	flag.StringVar(&cfg.HTTPConnStr, "http", cfg.HTTPConnStr, "http-connection")
	return cfg
}
