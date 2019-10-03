package utils

import (
	"log"
	"strings"

	"github.com/spf13/viper"
)

// GetConfig returns the global config
func GetConfig() *viper.Viper {
	c := viper.New()
	c.SetConfigType("yaml")
	c.SetConfigName("config")
	c.AddConfigPath(".")
	c.AutomaticEnv()

	c.SetDefault("debug", true)
	c.SetDefault("admins", []interface{}{})

	c.SetDefault("redis.host", "localhost:6379")
	c.SetDefault("redis.db", 0)
	c.SetDefault("redis.password", "")

	c.SetDefault("crisp.identifier", "")
	c.SetDefault("crisp.key", "")

	c.SetDefault("telegram.key", "")

	replacer := strings.NewReplacer(".", "_")
	c.SetEnvKeyReplacer(replacer)

	if err := c.ReadInConfig(); err != nil {
		log.Fatal(err.Error())
	}

	return c
}
