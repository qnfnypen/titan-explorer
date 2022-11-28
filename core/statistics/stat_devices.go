package statistics

import (
	"context"
	"time"
)

const DKeyUpdateDeviceInfo = "dk_update_device_info"

func (s *Statistic) UpdateDeviceInfo() {
	s.Once(context.Background(), DKeyUpdateDeviceInfo, func() error {
		log.Info("start update device info")
		start := time.Now()
		defer func() {
			log.Infof("update device info done, cost: %v", time.Since(start))
		}()

		time.Sleep(1 * time.Second)
		return nil
	})
}
