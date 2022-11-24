package statistics

import (
	"time"
)

func (s *Statistic) UpdateDeviceInfo() {
	log.Info("start update device info")
	start := time.Now()
	defer func() {
		log.Infof("update device info done, cost: %v", time.Since(start))
	}()

	log.Infof("doing")
	time.Sleep(2 * time.Second)
}
