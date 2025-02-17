package order

import (
	"github.com/gnasnik/titan-explorer/core"
)

const (
	baseCPUPrice     = 100
	baseRAMPrice     = 100
	baseStoragePrice = 2
)

func (m *Mgr) calculateCPUPrice(cores int) int {
	switch {
	case cores >= 1 && cores <= 4:
		return cores * baseCPUPrice * 1
	case cores >= 5 && cores <= 8:
		return (cores * baseCPUPrice * 9) / 10
	case cores >= 9 && cores <= 16:
		return (cores * baseCPUPrice * 8) / 10
	case cores >= 17 && cores <= 32:
		return (cores * baseCPUPrice * 7) / 10
	default:
		return 0
	}
}

func (m *Mgr) calculateRAMPrice(size int) int {
	switch {
	case size >= 1 && size <= 4:
		return size * baseRAMPrice * 1
	case size >= 5 && size <= 16:
		return (size * baseRAMPrice * 9) / 10
	case size >= 17 && size <= 32:
		return (size * baseRAMPrice * 8) / 10
	case size >= 33 && size <= 64:
		return (size * baseRAMPrice * 7) / 10
	default:
		return 0
	}
}

func (m *Mgr) calculateStoragePrice(size int) int {
	switch {
	case size >= 40 && size <= 100:
		return size * baseStoragePrice * 1
	case size >= 101 && size <= 500:
		return (size * baseStoragePrice * 8) / 10
	case size >= 501 && size <= 2000:
		return (size * baseStoragePrice * 6) / 10
	case size >= 2001 && size <= 4000:
		return (size * baseStoragePrice * 5) / 10
	default:
		return 0
	}
}

func (m *Mgr) calculateDurationCoefficient(hours int) int {
	switch {
	case hours >= 1 && hours <= 24:
		return 10
	case hours >= 25 && hours <= 72:
		return 9
	case hours >= 73 && hours <= 168:
		return 8
	case hours >= 169 && hours <= 720:
		return 7
	default:
		return 0
	}
}

// CalculateTotalCost calculates the total cost based on the order request configuration.
func (m *Mgr) CalculateTotalCost(config *core.OrderInfoReq) int {
	cpuCost := m.calculateCPUPrice(config.CPUCores)
	ramCost := m.calculateRAMPrice(config.RAMSize)
	storageCost := m.calculateStoragePrice(config.StorageSize)

	hourlyBaseCost := cpuCost + ramCost + storageCost
	durationCoefficient := m.calculateDurationCoefficient(config.Duration)

	totalCost := (hourlyBaseCost * config.Duration * durationCoefficient) / 10

	log.Infof("Total cost for the configuration: %d coins cpu:[%d] ram:[%d] storage:[%d] duration:[%d]", totalCost, cpuCost, ramCost, storageCost, durationCoefficient)

	// Detailed breakdown
	// log.Infoln("\nDetailed Breakdown:\n")
	// log.Infof("CPU Cost (per hour): %d\n", cpuCost)
	// log.Infof("RAM Cost (per hour): %d\n", ramCost)
	// log.Infof("Storage Cost (per hour): %d\n", storageCost)
	// log.Infof("Duration Coefficient: %d\n", durationCoefficient)
	// log.Infof("Total Duration: %d hours\n", config.Duration)

	return totalCost // Round to 2 decimal places
}
