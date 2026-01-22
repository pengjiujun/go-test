package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"strings"
)

type DBConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Database string
}

type JwtConfig struct {
	Secret        string
	Issuer        string
	ExpireSeconds int64
}

type LogConfig struct {
	Level      string
	Format     string
	Output     string
	FilePath   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
	Compress   bool
}

type Config struct {
	Database DBConfig
	Jwt      JwtConfig
	Log      LogConfig
}

var Conf *Config

func Load() {

	err := godotenv.Load(".env")
	if err != nil {
		panic("Error loading .env file")
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".") // 程序当前运行的目录
	v.AutomaticEnv()
	v.SetEnvPrefix("APP")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		panic(err)
	}
	Conf = &config
}
