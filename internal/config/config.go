package config

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	OAuth2 struct {
		Casdoor struct {
			BaseURL          string `mapstructure:"base_url"`
			ClientID         string `mapstructure:"client_id"`
			ClientSecret     string `mapstructure:"client_secret"`
			RedirectURI      string `mapstructure:"redirect_uri"`
			OrganizationName string `mapstructure:"organization_name"`
			ApplicationName  string `mapstructure:"application_name"`
			Cert             string `mapstructure:"cert"`
		} `mapstructure:"casdoor"`
	} `mapstructure:"oauth2"`
	JWT struct {
		Secret             string `mapstructure:"secret"`
		AccessTokenExpiry  string `mapstructure:"access_token_expiry"`
		RefreshTokenExpiry string `mapstructure:"refresh_token_expiry"`
		Issuer             string `mapstructure:"issuer"`
	} `mapstructure:"jwt"`
	Database struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"database"`
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Kafka struct {
		Brokers []string `mapstructure:"brokers"`
		Topic   string   `mapstructure:"topic"`
	} `mapstructure:"kafka"`
	Session struct {
		MaxConcurrentSessions int           `mapstructure:"max_concurrent_sessions"`
		CleanupInterval       time.Duration `mapstructure:"cleanup_interval"`
	} `mapstructure:"session"`
	Security struct {
		RateLimit struct {
			LoginAttempts int           `mapstructure:"login_attempts"`
			Window        time.Duration `mapstructure:"window"`
		} `mapstructure:"rate_limit"`
		Cookie struct {
			Secure   bool   `mapstructure:"secure"`
			SameSite string `mapstructure:"same_site"`
			HttpOnly bool   `mapstructure:"http_only"`
		} `mapstructure:"cookie"`
	} `mapstructure:"security"`
}

func Load() (*Config, error) {
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func NewCasdoorClient(cfg *Config) *casdoorsdk.Client {
	// cert is file, config should be a path to the certificate file
	if cfg.OAuth2.Casdoor.Cert != "" {

	}
	client := casdoorsdk.NewClient(
		cfg.OAuth2.Casdoor.BaseURL,
		cfg.OAuth2.Casdoor.ClientID,
		cfg.OAuth2.Casdoor.ClientSecret,
		cfg.OAuth2.Casdoor.RedirectURI,
		cfg.OAuth2.Casdoor.OrganizationName,
		cfg.OAuth2.Casdoor.ApplicationName,
	)

	return client
}
