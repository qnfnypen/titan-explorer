package sync

import (
	"context"
	"log"

	"github.com/gnasnik/titan-explorer/api"
)

var (
	ctx = context.Background()
)

func main() {
	scli, _ := api.GetSchedulerClient(ctx, "Asia-China-Guangdong-Shenzhen")

	cid := "bafybeieq4ahq3nmbzfkxwml6kys5r4jaqwdl3swuv5vzie4upo2sgrl4ce"
	unSyncAids := []string{"Europe-Germany-Hesse-FrankfurtamMain", "NorthAmerica-UnitedStates"}
	cids, err := api.SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(cids)
	}

	scli, _ = api.GetSchedulerClient(ctx, "Asia-Singapore")
	cid = "bafybeibry7lqb5soj52vl77fqp2wigbnwrklwaa5w77y2tvsthksldymsa"
	unSyncAids = []string{"Europe-Germany-Hesse-FrankfurtamMain", "NorthAmerica-UnitedStates", "Asia-Vietnam-Hanoi-Hanoi"}
	cids, err = api.SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(cids)
	}

	scli, _ = api.GetSchedulerClient(ctx, "Asia-Japan-Tokyo-Tokyo")
	cid = "bafybeibz4nj72svea2goowncunmmukt3q67kfw4tvud52unkiutifpy5du"
	unSyncAids = []string{"Asia-HongKong","Asia-Vietnam-Hanoi-Hanoi","NorthAmerica-UnitedStates-California"}
	cids, err = api.SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	if err != nil {
		log.Fatal(err)
	} else {
		log.Println(cids)
	}
}
