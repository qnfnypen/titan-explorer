package job

import (
	"context"
	"testing"

	"github.com/gnasnik/titan-explorer/api"
)

func TestSyncAsset(t *testing.T) {
	var (
		ctx = context.Background()

		cid           = "bafybeic6sx6kvnndanrzlddltcpw7nerjfkahfxzj536die4nmojwlnmqi"
		owner         = "titan17ljevhtqu4vx6y7k743jyca0w8gyfu2466e8x3"
		areaID        = "NorthAmerica-UnitedStates"
		unSyncAreaIDs = []string{"Asia-China-Guangdong-Shenzhen", "Asia-HongKong",
			"Asia-Japan-Tokyo-Tokyo", "Asia-Singapore", "Asia-SouthKorea-Seoul-Seoul",
			"Asia-Vietnam-Hanoi-Hanoi", "Europe-Germany-Hesse-FrankfurtamMain",
			"Europe-UnitedKingdom-England-London", "NorthAmerica-Canada"}
	)

	scli, err := api.GetSchedulerClient(ctx, areaID)
	if err != nil {
		t.Fatalf("get client of scheduler error:%v", err)
	}

	aids, err := SyncShedulers(ctx, scli, cid, 0, owner, unSyncAreaIDs)
	if err != nil {
		t.Fatalf("sync asset error:%v", err)
	}

	t.Log(aids)
}
