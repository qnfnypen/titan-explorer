package main

import (
	"fmt"
	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/viper"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	OsSignal := make(chan os.Signal, 1)

	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("toml")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("reading config file: %v\n", err)
	}

	var cfg config.Config
	if err := viper.Unmarshal(&cfg); err != nil {
		log.Fatalf("unmarshaling config file: %v\n", err)
	}

	if cfg.Mode == "debug" {
		logging.SetDebugLogging()
	}

	if err := core.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	go api.RunTask()
	srv, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("create api server: %v\n", err)
	}
	go srv.Run()

	signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM)
	_ = <-OsSignal
	srv.Close()

	fmt.Printf("Exiting received OsSignal\n")
}
