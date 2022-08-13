package repo

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigReader struct {
	AdminID         int64
	BotToken        string
	PollingInterval int64
	DB              string
	VPNHost         string
	CardNumber      string
	Message         MessageReader
}

type MessageReader struct {
	Start string
	Apps  string
	Setup string
	Info  string
	Pay   string
}

var Config ConfigReader = InitConfig()

func InitConfig() ConfigReader {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Error(err)
		panic(err)
	}
	config := ConfigReader{}
	viper.Unmarshal(&config)
	return config
}
