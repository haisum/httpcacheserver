package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/haisum/httpcacheserver/proxy"
	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("HTTP_PORT", 8081)
	viper.SetDefault("HTTP_HOST", "")
	viper.SetDefault("PROXY_URL", "http://mirror.centos.org/centos/")
	viper.SetDefault("PROXY_SUFFIX", "/proxy")
	viper.SetDefault("DATA_DIR", "./cache")
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.WithError(err).Warn("Couldn't read config file.")
	}
	viper.SetEnvPrefix("HCS")
	viper.AutomaticEnv()
}

func main() {
	err := proxy.Start(viper.GetString("PROXY_URL"),
		viper.GetString("PROXY_SUFFIX"),
		viper.GetString("DATA_DIR"),
		viper.GetString("HTTP_HOST"),
		viper.GetInt("HTTP_PORT"))
	if err != nil {
		log.WithError(err).Fatal("Error happened in starting proxy server.")
	}
}
