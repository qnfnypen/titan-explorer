package statistics

import (
	"time"
)

func (s *Statistic) UpdateDeviceInfo() error {
	log.Info("start update device info")
	start := time.Now()
	defer func() {
		log.Infof("update device info done, cost: %v", time.Since(start))
	}()

	time.Sleep(1 * time.Second)
	return nil
}
