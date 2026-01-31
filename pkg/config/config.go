package config

import (
	"github.com/fsnotify/fsnotify"
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

type RedisConfig struct {
	Port     int
	Addr     string
	Password string
	DB       int
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
	Redis    RedisConfig
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

	// 初始加载配置
	if err := v.Unmarshal(&Conf); err != nil {
		panic(err)
	}

	// 监听配置文件变化
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		// 重载配置到全局变量 Conf
		if err := v.Unmarshal(&Conf); err != nil {
			panic(err)
		}
	})
}
