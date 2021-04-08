package test

import "github.com/caarlos0/env"

type config struct {
	PGHost     string `env:"PG_HOST,required"`
	PGPort     uint   `env:"PG_PORT" envDefault:"5432"`
	PGName     string `env:"PG_NAME,required"`
	PGUser     string `env:"PG_USER,required"`
	PGPassword string `env:"PG_PASSWORD,required"`
	PGUseSSL   bool   `env:"PG_USE_SSL" envDefault:"false"`
}

func parseEnv() (*config, error) {
	var cfg config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
