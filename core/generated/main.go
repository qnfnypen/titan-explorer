package main

import (
	"github.com/gnasnik/titan-explorer/config"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gen"
	"gorm.io/gorm"
	"log"
)

func main() {
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

	g := gen.NewGenerator(gen.Config{
		OutPath: "core/generated/query",
		Mode:    gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface,
	})

	db, _ := gorm.Open(mysql.Open(cfg.DatabaseURL), &gorm.Config{})
	g.UseDB(db)

	users := g.GenerateModel("users")
	loginLog := g.GenerateModel("login_log")
	operationLog := g.GenerateModel("operation_log")
	schedulers := g.GenerateModel("schedulers")
	deviceInfo := g.GenerateModel("device_info")
	hoursDaily := g.GenerateModel("hour_daily")
	incomeDaily := g.GenerateModel("income_daily")
	retrievalInfo := g.GenerateModel("retrieval_info")
	taskInfo := g.GenerateModel("task_info")

	g.ApplyBasic(users)
	g.ApplyBasic(loginLog)
	g.ApplyBasic(operationLog)
	g.ApplyBasic(schedulers)
	g.ApplyBasic(deviceInfo)
	g.ApplyBasic(hoursDaily)
	g.ApplyBasic(incomeDaily)
	g.ApplyBasic(retrievalInfo)
	g.ApplyBasic(taskInfo)

	g.Execute()
}
