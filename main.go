package main

import (
	"github.com/robrotheram/gogallery/config"
	"github.com/robrotheram/gogallery/datastore"
	"github.com/robrotheram/gogallery/web"
	"github.com/robrotheram/gogallery/worker"
	"github.com/spf13/viper"
	"log"
	"strings"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.SetEnvPrefix("GLLRY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := viper.Unmarshal(&config.Config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	log.Println(config.Config.Database)
	worker.StartWorkers()

	datastore.Cache = datastore.NewDataStore()

	go func() {
		datastore.ScanPath(config.Config.Gallery.Basepath)
	}()
	web.Serve()
}
