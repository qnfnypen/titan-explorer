package api

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gnasnik/titan-explorer/core/storage"
)

var (
	ctx = context.Background()
)

func TestSyncShedulers(t *testing.T) {
	_, err := GetSchedulerClient(ctx, "NorthAmerica-Canada")
	if err != nil {
		t.Fatal(err)
	}

	// cid := "bafybeiecvk3yk3qq6iyn5rj3s5rt3zsb3vuxrh2g527j5aihszicj2sdtu"
	// unSyncAids := []string{"Asia-Vietnam-Hanoi-Hanoi"}
	// cids, err := SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	// t.Log(cids)

	// scli, _ = GetSchedulerClient(ctx, "Asia-Singapore")
	// cid = "bafybeibry7lqb5soj52vl77fqp2wigbnwrklwaa5w77y2tvsthksldymsa"
	// unSyncAids = []string{"Europe-Germany-Hesse-FrankfurtamMain", "NorthAmerica-UnitedStates", "Asia-Vietnam-Hanoi-Hanoi"}
	// cids, err = SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// t.Log(cids)

	// scli, _ = GetSchedulerClient(ctx, "Asia-Japan-Tokyo-Tokyo")
	// cid = "bafybeibz4nj72svea2goowncunmmukt3q67kfw4tvud52unkiutifpy5du"
	// unSyncAids = []string{"Asia-HongKong", "Asia-Vietnam-Hanoi-Hanoi", "NorthAmerica-UnitedStates-California"}
	// cids, err = SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// t.Log(cids)
}

func TestParseHost(t *testing.T) {
	uri := "https://test26-scheduler.titannet.io:3456/rpc/v0"

	aurl, err := url.Parse(uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(aurl.Host)

	uri, _, _ = strings.Cut(aurl.Host, ":")
	ips, err := net.LookupIP(uri)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ips[0].String())

	for _, v := range ips {
		t.Log(v.String())
	}
}

func TestMoveHash(t *testing.T) {

	url := "https://storage-test.titannet.io/api/v1/storage/move_node"
	method := "POST"

	// Asia-China-Guangdong-Shenzhen
	payload := strings.NewReader(`{
	  "from_area_id":"Asia-China-Guangdong-Shenzhen",
	  "node_id":"c_7c7bd97b-6742-4f74-abd6-5b2c2a4b2744",
	  "to_area_id":"NorthAmerica-UnitedStates"}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Jwtauthorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI2NTA1NjYsImlkIjoibTEyNTY2Njg3MjVAZ21haWwuY29tIiwib3JpZ19pYXQiOjE3MjI1NjQxNjYsInJvbGUiOjB9.6Ow4tDNx5ga47hc9RbU4X4PELW_eNEmJihSthLryrnw")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal()
	}
	t.Log(string(body))
}

func TestUploadTf(t *testing.T) {
	url := "https://storage-test.titannet.io/api/v1/storage/temp_file/upload"
	method := "POST"

	payload := strings.NewReader(`{
    "area_ids": ["Asia-China-Guangdong-Shenzhen", "NorthAmerica-UnitedStates-Ohio-Columbus"],
    "asset_cid": "bafkreih2b6puzcjebqijcawavwh2kjucldk7rfubx2xlku67wgcp7mmh4e",
    "asset_name": "icon_1.png",
    "asset_size": 73830
}`)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		t.Fatal(err)
	}
	req.Header.Add("Jwtauthorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MjI1NjkxNjksImlkIjoibTEyNTY2Njg3MjVAZ21haWwuY29tIiwib3JpZ19pYXQiOjE3MjI0ODI3NjksInJvbGUiOjB9.hpzoIH7mxGy4CMFmDDpGmT0ig6RSWL9KVUOHhm2xDZY")
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(body))
}

func TestMove(t *testing.T) {
	os.Setenv("ETCD_USERNAME", "web")
	os.Setenv("ETCD_PASSWORD", "web_123")
	var req MoveNodeReq
	req.FromAreaID = "Asia-SouthKorea-Seoul-Seoul"
	req.NodeID = "c_7ceddfd2-ecab-4b0f-96de-b0169d00c9b5"
	req.ToAreaID = "Asia-Vietnam-Hanoi-Hanoi"
	// 将node节点从from area移出
	fscli, err := getSchedulerClient(ctx, req.FromAreaID)
	if err != nil {
		t.Fatal(err)
	}
	info, err := fscli.MigrateNodeOut(ctx, req.NodeID)
	if err != nil {
		t.Fatal(err)
	}
	tscli, err := getSchedulerClient(ctx, req.ToAreaID)
	if err != nil {
		t.Fatal(err)
	}
	err = tscli.MigrateNodeIn(ctx, info)
	if err != nil {
		t.Fatal(err)
	}
	err = fscli.CleanupNode(ctx, req.NodeID, info.Key)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("success")
}

func TestChange(t *testing.T) {
	cid := "bafkreicfnhjjtgihpix2ec2mnz7p6k2kh5ra73eij4sk5hhzrvvlkfmx4e"
	hash, err := storage.CIDToHash(cid)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hash)
}
