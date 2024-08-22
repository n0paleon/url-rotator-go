package config

import (
	"github.com/spf13/viper"
)

var (
	V *viper.Viper
)

func InitConfig(filename, filepath string) *viper.Viper {
	cfg := viper.New()

	cfg.SetEnvPrefix("VP")
	cfg.AutomaticEnv()

	cfg.SetConfigFile(filename)
	cfg.AddConfigPath(filepath)
	if err := cfg.ReadInConfig(); err != nil {
		panic(err)
	}

	V = cfg

	return cfg
}
