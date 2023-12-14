package main

import (
	"flag"
	"github.com/gnasnik/titan-explorer/api"
	logging "github.com/ipfs/go-log/v2"
)

var (
	etcd     string
	user     string
	password string
	token    string
)

func init() {
	flag.StringVar(&etcd, "etcd", "", "etcd address")
	flag.StringVar(&user, "user", "", "etcd user")
	flag.StringVar(&password, "password", "", "etcd password")
	flag.StringVar(&token, "token", "", "storage api authenticate token")
}

func main() {
	flag.Parse()

	logging.SetDebugLogging()

	var address []string
	address = append(address, etcd)
	eClient, err := api.NewEtcdClient(address)
	if err != nil {
		log.Fatal("New etcdClient Failed: %v", err)
	}

	schedulers, err := api.FetchSchedulersFromEtcd(eClient)
	if err != nil {
		log.Fatal("fetch scheduler from etcd Failed: %v", err)
	}

	if len(schedulers) == 0 {
		log.Fatal("no scheduler found")
	}

	downloader := newDownloader(token, schedulers)
	go downloader.async()

	log.Infof("Started")
	downloader.run()
}
