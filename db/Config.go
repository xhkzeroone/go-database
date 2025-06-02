package db

import (
	"github.com/spf13/viper"
	"strings"
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
	}
}
