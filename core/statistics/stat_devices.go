package statistics

import (
	"context"
	"time"
)

const DKeyUpdateDeviceInfo = "dk_update_device_info"

func (s *Statistic) UpdateDeviceInfo() {
	s.once(context.Background(), DKeyUpdateDeviceInfo, time.Minute, func() error {
		log.Info("start update device info")
		start := time.Now()
		defer func() {
			log.Infof("update device info done, cost: %v", time.Since(start))
		}()

		log.Infof("doing")
		time.Sleep(2 * time.Second)
		return nil
	})
}
