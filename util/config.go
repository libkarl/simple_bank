package util

import "github.com/spf13/viper"

// Config stores all configuration of the application.
// The values are read by viper from a config file or enviroment variables.
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env") // it can be also set to json,xml or yaml for example
	// this will check if enviroment variables match something from
	// the existing keys in .env file, pokud je nalezena shoda jsou změny načteny do viperu
	// lze tedy před spuštěním serveru přes make server
	// přepsat libovolnou proměnnou, která je nadefinovaná v app.env
	// env SERVER_ADDRESS=0.0.0.0:8081 make server -> viper použije port 8081 místo 8080, který je nastavení v app.env
	// to se používá při nasazení do produkce, jelikož díky tomu lze server spustit v produkci s odlišnými
	// enviroment proměnnými a viper je prostě jen načte místo těch defaultních co mám lokálně
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
