package server

type ServerConfig struct {
	Environment string `env:"SERVER_ENVIRONMENT" envDefault:"local"`
	GogmConfig  GogmConfig
	WebConfig   WebConfig
}

type GogmConfig struct {
	Host     string `env:"GOGM_HOST" envDefault:"localhost"`
	Port     int    `env:"GOGM_PORT" envDefault:"7687"`
	PoolSize int    `env:"GOGM_POOL_SIZE" envDefault:"50"`
	Username string `env:"GOGM_USERNAME,required"`
	Password string `env:"GOGM_PASSWORD,required"`
}

type WebConfig struct {
	Timeout int `env:"WEB_TIMEOUT" envDefault:"60"`
	Port    int `env:"WEB_PORT" envDefault:"8080"`
}
