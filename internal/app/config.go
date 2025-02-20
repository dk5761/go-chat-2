package app

import "github.com/spf13/viper"

func loadConfig() error {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("yaml")   // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath("/app/config")
	viper.AddConfigPath("config")
	viper.AddConfigPath(".")

	// Set defaults
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.shutdown_timeout", "30s")

	return viper.ReadInConfig()
}
