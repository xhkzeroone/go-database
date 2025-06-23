package db

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Host     string `mapstructure:"host" yaml:"host"`
	Port     string `mapstructure:"port" yaml:"port"`
	User     string `mapstructure:"user" yaml:"user"`
	Password string `mapstructure:"password" yaml:"password"`
	DBName   string `mapstructure:"dbname" yaml:"dbname"`
	Schema   string `mapstructure:"schema" yaml:"schema"`
	SSLMode  string `mapstructure:"sslmode" yaml:"sslmode"`
	Debug    bool   `mapstructure:"debug" yaml:"debug"`
	Driver   string `mapstructure:"driver" yaml:"driver"`

	MaxOpenConns    int   `mapstructure:"max_open_conns" yaml:"max_open_conns"`
	MaxIdleConns    int   `mapstructure:"max_idle_conns" yaml:"max_idle_conns"`
	ConnMaxLifetime int64 `mapstructure:"conn_max_lifetime" yaml:"conn_max_lifetime"` // đơn vị giây
}

func DefaultConfig() *Config {
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "postgres")
	viper.SetDefault("database.schema", "public")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.debug", true)
	viper.SetDefault("database.driver", "postgres")

	viper.SetDefault("database.max_open_conns", 10)
	viper.SetDefault("database.max_idle_conns", 5)
	viper.SetDefault("database.conn_max_lifetime", 3600) // 1 giờ

	return &Config{
		Host:     viper.GetString("database.host"),
		Port:     viper.GetString("database.port"),
		User:     viper.GetString("database.user"),
		Password: viper.GetString("database.password"),
		DBName:   viper.GetString("database.dbname"),
		Schema:   viper.GetString("database.schema"),
		SSLMode:  viper.GetString("database.sslmode"),
		Debug:    viper.GetBool("database.debug"),
		Driver:   viper.GetString("database.driver"),

		MaxOpenConns:    viper.GetInt("database.max_open_conns"),
		MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
		ConnMaxLifetime: viper.GetInt64("database.conn_max_lifetime"),
	}
}
