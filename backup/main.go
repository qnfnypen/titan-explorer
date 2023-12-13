package main

import (
	"flag"
	"github.com/gnasnik/titan-explorer/api"
)

var (
	etcd     string
	user     string
	password string
	listen   string
	token    string
)

func init() {
	flag.StringVar(&etcd, "etcd", "", "etcd address")
	flag.StringVar(&user, "user", "", "etcd user")
	flag.StringVar(&password, "password", "", "etcd password")
	flag.StringVar(&token, "token", "", "storage api authenticate token")
	flag.StringVar(&listen, "listen", ":3001", "api listen address")
}

func main() {
	flag.Parse()

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

	downloader := newDownloader(token, schedulers)
	go downloader.async()

	downloader.run()
}
