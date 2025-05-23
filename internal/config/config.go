package config

import (
	"github.com/caarlos0/env/v6"
)

type Config struct {
	Port               int    `env:"PORT" envDefault:"8080"`
	GRPCPort           int    `env:"GRPC_PORT" envDefault:"8090"`
	AdditionMs         int    `env:"ADDITION_MS" envDefault:"1000"`
	SubtractionMs      int    `env:"SUBTRACTION_MS" envDefault:"1000"`
	MultiplicationMs   int    `env:"MULTIPLICATION_MS" envDefault:"1500"`
	DivisionMs         int    `env:"DIVISION_MS" envDefault:"2000"`
	ComputingPower     int    `env:"COMPUTING_POWER" envDefault:"3"`
	AgentPeriodicityMs int    `env:"AGENT_PERIODICITY_MS" envDefault:"500"`
	DBPath             string `env:"DB_PATH" envDefault:"./data/calculator.db"`
	GRPCHost           string `env:"GRPC_HOST" envDefault:"localhost"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
