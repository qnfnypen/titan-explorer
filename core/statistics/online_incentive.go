package statistics

import (
	"context"
	"github.com/gnasnik/titan-explorer/core/dao"
	"github.com/golang-module/carbon/v2"
	"time"
)

func runGenOnlineIncentive() error {
	log.Info("Start to generate device online incentive")
	start := time.Now()
	defer func() {
		log.Infof("generate device online incentive done, cost: %v", time.Since(start))
	}()

	ctx := context.Background()
	yesterday := carbon.Yesterday().StartOfDay().StdTime()

	isGenerated, err := dao.IsGeneratedOnlineIncentive(ctx, yesterday)
	if err != nil {
		log.Errorf("IsGeneratedOnlineIncentive: %v", err)
		return err
	}

	if isGenerated {
		return nil
	}

	err = dao.GenerateEligibleOnlineDevices(ctx)
	if err != nil {
		log.Errorf("GenerateEligibleOnlineDevices: %v", err)
	}

	// 计算今天的, 昨天的, 七天的, 30天的在线收益

	return nil
}
