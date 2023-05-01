package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBUri                 string        `mapstructure:"MONGODB_LOCAL_URI"`
	RedisUri              string        `mapstructure:"REDIS_URL"`
	Port                  string        `mapstructure:"PORT"`
	EmailFrom             string        `mapstructure:"EMAIL_FROM"`
	Origin                string        `mapstructure:"CLIENT_ORIGIN"`
	PrivBuf               []byte        `mapstructure:"-"`
	PubBuf                []byte        `mapstructure:"-"`
	AccessTokenExpiresIn  time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN"`
	RefreshTokenExpiresIn time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN"`
	AccessTokenMaxAge     int           `mapstructure:"ACCESS_TOKEN_MAXAGE"`
	RefreshTokenMaxAge    int           `mapstructure:"REFRESH_TOKEN_MAXAGE"`

	SMTPHost          string `mapstructure:"SMTP_HOST"`
	SMTPPass          string `mapstructure:"SMTP_PASS"`
	SMTPPort          int    `mapstructure:"SMTP_PORT"`
	SMTPUser          string `mapstructure:"SMTP_USER"`
	GrpcServerAddress string `mapstructure:"GRPC_SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config,
	err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName("config")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
