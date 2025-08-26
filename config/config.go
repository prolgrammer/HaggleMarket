package config

import (
	"fmt"
	"log"
	"time"

	flag "github.com/spf13/pflag"

	"github.com/spf13/viper"
)

var (
	EnvDev  = "dev"
	EnvTest = "test"
	EnvProd = "prod"
)

type DBConfig struct {
	Host         string
	User         string
	Password     string
	DefAdmPswd   string
	Name         string
	Port         string
	MaxOpenConns int
	MaxIdleConns int
}

func (dbc *DBConfig) GetDBConnString() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbc.Host, dbc.User, dbc.Password, dbc.Name, dbc.Port,
	)
}

type ServerConfig struct {
	Domain  string
	Port    string
	GinMode string
}

type TokenConfig struct {
	Secret               string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
}

type EmailConfig struct {
	Enabled    bool
	From       string
	Password   string
	SMTPHost   string
	SMTPPort   int
	WorkerPool int
	QueueSize  int
}

type AuditConfig struct {
	WorkerPoolSize int
	QueueSize      int
	BatchSize      int
}

type Config struct {
	AccessKey     string
	SecretKey     string
	BucketName    string
	URL           string
	SigningRegion string
	Environment   string

	DB     DBConfig
	Server ServerConfig
	Token  TokenConfig
	Email  EmailConfig
	Audit  AuditConfig
}

func LoadConfig() *Config {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config reading failed: %v", err)
	}

	if err := viper.BindPFlag("mail", flag.Lookup("mail")); err != nil {
		log.Printf("Binding mail flag failed: %v", err)
	}

	// Дефолтные значения для конфига базы данных
	viper.SetDefault("DB_PORT", 5432)
	viper.SetDefault("DB_MAX_OPEN_CONNS", 25)
	viper.SetDefault("DB_MAX_IDLE_CONNS", 10)
	viper.SetDefault("SERVER_PORT", 8080)
	viper.SetDefault("AUDIT_BATCH_SIZE", 100)

	config := &Config{
		AccessKey:     viper.GetString("ACCESS_KEY"),
		SecretKey:     viper.GetString("SECRET_KEY"),
		BucketName:    viper.GetString("BUCKET_NAME"),
		URL:           viper.GetString("URL"),
		SigningRegion: viper.GetString("SIGNING_REGION"),
		Environment:   viper.GetString("ENVIRONMENT"),
		DB: DBConfig{
			Host:         viper.GetString("DB_HOST"),
			User:         viper.GetString("DB_USER"),
			Password:     viper.GetString("DB_PASSWORD"),
			DefAdmPswd:   viper.GetString("DEFAULT_ADMIN_PSWD"),
			Name:         viper.GetString("DB_NAME"),
			Port:         viper.GetString("DB_PORT"),
			MaxOpenConns: viper.GetInt("DB_MAX_OPEN_CONNS"),
			MaxIdleConns: viper.GetInt("DB_MAX_IDLE_CONNS"),
		},
		Server: ServerConfig{
			Domain:  viper.GetString("SERVER_DOMAIN"),
			Port:    viper.GetString("SERVER_PORT"),
			GinMode: viper.GetString("GIN_MODE"),
		},
		Token: TokenConfig{
			Secret:               viper.GetString("TOKEN_SECRET"),
			AccessTokenDuration:  viper.GetDuration("TOKEN_ACCESS_DURATION"),
			RefreshTokenDuration: viper.GetDuration("TOKEN_REFRESH_DURATION"),
		},
		Email: EmailConfig{
			Enabled:    viper.GetBool("EMAIL_ENABLED") || viper.GetBool("mail"),
			From:       viper.GetString("FROM_EMAIL"),
			Password:   viper.GetString("FROM_PASSWORD"),
			SMTPHost:   viper.GetString("SMTP_HOST"),
			SMTPPort:   viper.GetInt("SMTP_PORT"),
			WorkerPool: viper.GetInt("EMAIL_WORKER_POOL"),
			QueueSize:  viper.GetInt("EMAIL_QUEUE_SIZE"),
		},
		Audit: AuditConfig{
			WorkerPoolSize: viper.GetInt("AUDIT_WORKER_POOL"),
			QueueSize:      viper.GetInt("AUDIT_QUEUE_SIZE"),
			BatchSize:      viper.GetInt("AUDIT_BATCH_SIZE"),
		},
	}

	return config
}
