package api

import (
	"context"
	"testing"
)

var (
	ctx = context.Background()
)

func TestSyncShedulers(t *testing.T) {
	scli, _ := GetSchedulerClient(ctx, "NorthAmerica-Canada")

	cid := "bafybeiecvk3yk3qq6iyn5rj3s5rt3zsb3vuxrh2g527j5aihszicj2sdtu"
	unSyncAids := []string{"Asia-Vietnam-Hanoi-Hanoi"}
	cids, err := SyncShedulers(ctx, scli, "", cid, 0, unSyncAids)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(cids)

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
