package util

import "github.com/spf13/viper"

func LoadConfig(path string) error {
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AddConfigPath(path)
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	return err
}