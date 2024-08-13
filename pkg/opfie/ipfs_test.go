package opfie

import (
	"context"
	"testing"
)

var (
	cli    *IPFSClient
	ctx    = context.Background()
	cidStr = "bafkreidtuosuw37f5xmn65b3ksdiikajy7pwjjslzj2lxxz2vc4wdy3zku"
)

func TestMain(m *testing.M) {
	var err error
	url := "/ip4/39.108.214.29/tcp/5001"

	cli, err = NewIPFSClient(url)
	if err != nil {
		panic(err)
	}

	m.Run()
}

func TestAddFileByCID(t *testing.T) {
	cidStr = "QmXeKtRSz7SVKp8Qh6tXtv6whRF5wxwQPwsgA38houakhz"
	err := cli.AddFileByCID(ctx, cidStr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetInfoByCID(t *testing.T) {
	cidStr = "bafybeicpifbyoqndom3usp55ojqu4mei3anhusgto2fpyyerftlfkzcuky"
	links, size, err := cli.GetInfoByCID(ctx, cidStr)
	if err != nil {
		t.Fatal(err)
	}

	for _, v := range links {
		t.Log(*v)
	}

	t.Log(size)
}
