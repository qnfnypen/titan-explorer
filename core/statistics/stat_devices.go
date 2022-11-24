package statistics

import (
	"context"
	"time"
)

const DKeyUpdateDeviceInfo = "dk_update_device_info"
const LockerTTL = 30 * time.Second

func (s *Statistic) UpdateDeviceInfo() {
	s.Once(context.Background(), DKeyUpdateDeviceInfo, LockerTTL, func() error {
		log.Info("start update device info")
		start := time.Now()
		defer func() {
			log.Infof("update device info done, cost: %v", time.Since(start))
		}()

		time.Sleep(1 * time.Second)
		return nil
	})
}
