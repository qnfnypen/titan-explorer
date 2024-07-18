package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gnasnik/titan-explorer/api"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/gnasnik/titan-explorer/core/oplog"
	"github.com/gnasnik/titan-explorer/pkg/oss"
	logging "github.com/ipfs/go-log/v2"
	"github.com/spf13/viper"
)

// @title Titan Explorer API
// @version 1.0
// @description This is titan explorer backend server.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
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
	config.Cfg = cfg
	if cfg.Mode == "debug" {
		logging.SetDebugLogging()
	}

	if err := dao.Init(&cfg); err != nil {
		log.Fatalf("initital: %v\n", err)
	}

	if err := oss.InitFromCfg(cfg.Oss); err != nil {
		log.Fatalf("init oss: %v\n", err)
	}

	oplog.Subscribe(context.Background())

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
