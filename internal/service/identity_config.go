package service

import (
	"os"
	"strconv"
)

type ConfigOpts func(IdentityConfig) IdentityConfig

type IdentityConfig struct {
	host string
	port int
	user string
	pass string
}

func WithHost(host string) ConfigOpts {
	return func(cfg IdentityConfig) IdentityConfig {
		cfg.host = host
		return cfg
	}
}

func WithPort(port int) ConfigOpts {
	return func(cfg IdentityConfig) IdentityConfig {
		cfg.port = port
		return cfg
	}
}

func WithUser(user string) ConfigOpts {
	return func(cfg IdentityConfig) IdentityConfig {
		cfg.user = user
		return cfg
	}
}

func WithPass(pass string) ConfigOpts {
	return func(cfg IdentityConfig) IdentityConfig {
		cfg.pass = pass
		return cfg
	}
}

func NewIdentityConfig(opts ...ConfigOpts) IdentityConfig {
	cfg := IdentityConfig{
		host: "127.0.0.1",
		port: 8080,
		user: "John",
		pass: "VMw@re1!",
	}

	//read host from env
	host := os.Getenv("IDM_HOST")
	if host != "" {
		cfg.host = host
	}

	//read port from env
	port := os.Getenv("IDM_PORT")
	if port != "" {
		cfg.port, _ = strconv.Atoi(port)
	}

	//read user from env
	user := os.Getenv("IDM_USER")
	if user != "" {
		cfg.user = user
	}

	//read pass from env
	pass := os.Getenv("IDM_PASS")
	if pass != "" {
		cfg.pass = pass
	}

	for _, opt := range opts {
		cfg = opt(cfg)
	}

	return cfg
}
