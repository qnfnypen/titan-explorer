package order

import (
	"context"
	"time"

	zlog "log"

	"github.com/gnasnik/titan-explorer/core"
	"github.com/gnasnik/titan-explorer/core/chain"
	"github.com/gnasnik/titan-explorer/core/dao"
	kub "github.com/gnasnik/titan-explorer/core/kubesphere"
	"github.com/gnasnik/titan-explorer/core/oprds"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("order")

const (
	timeInterval  = 10 * time.Second
	timeInterval2 = 30 * time.Minute
	orderExecLock = "container_platform_order_exec_lock"
)

// Mgr manages order resources.
type Mgr struct {
	mDB      *dao.Mgr
	kubMgr   *kub.Mgr
	chainMgr *chain.Mgr
}

// NewOrderManager creates a new instance of Mgr for managing orders.
func NewOrderManager(db *dao.Mgr, k *kub.Mgr, c *chain.Mgr) *Mgr {
	m := &Mgr{}

	m.mDB = db
	m.kubMgr = k
	m.chainMgr = c

	res, err := oprds.GetClient().RedisClient().SetNX(context.Background(), orderExecLock, "1", 50*time.Second).Result()
	if err != nil || !res {
		return m
	}

	zlog.Println("get container_platform_order_exec_lock exec")
	go m.startTimer()
	go m.startTimer2()

	return m
}

// CheckChainMgr 检测 chain mgr
func (m *Mgr) CheckChainMgr() error {
	_, err := m.chainMgr.GetBalance("titan1u5vpfzh3eruy07rdx4884kjxgpqsxgg30ekum7")

	return err
}

func (m *Mgr) startTimer() {
	ticker := time.NewTicker(timeInterval)
	defer ticker.Stop()

	for {
		<-ticker.C

		m.checkOrderPaid()
		m.createSpaceFromOrders()
		m.checkOrderRenewal()

		m.checkOrderUpgrade()   // upgrade new order
		m.checkOrderAbandoned() // upgrade old order
	}
}

func (m *Mgr) startTimer2() {
	ticker := time.NewTicker(timeInterval2)
	defer ticker.Stop()

	for {
		<-ticker.C

		m.deleteOrders()
		m.checkOrderExpired()
		m.checkOrderActive()
		// cleanOrders()
	}
}

func (m *Mgr) checkOrderPaid() {
	ids, err := m.mDB.LoadOrderIDsByStatus(core.OrderStatusCreated)
	if err != nil {
		log.Errorf("checkOrderPaid LoadOrdersByStatus err:%s", err.Error())
		return
	}

	if len(ids) == 0 {
		return
	}

	tList, err := m.chainMgr.GetOrders(ids)
	if err != nil {
		return
	}

	for _, tOrder := range tList {
		if tOrder.Status != chain.Active {
			continue
		}

		dOrder, err := m.mDB.LoadOrderByID(tOrder.ID)
		if err != nil {
			log.Errorf("checkOrderPaid LoadOrderByID %s err:%s", dOrder.ID, err.Error())
			continue
		}

		hourDuration := int(tOrder.Duration) / 600

		dOrder.CPUCores = int(tOrder.Resource.CPU)
		dOrder.RAMSize = int(tOrder.Resource.Memory)
		dOrder.StorageSize = int(tOrder.Resource.Disk)
		dOrder.Duration = hourDuration
		dOrder.ExpiredAt = dOrder.CreatedAt.Add(time.Hour * time.Duration(dOrder.Duration))
		dOrder.Price = int(tOrder.LockedFunds)
		dOrder.Status = core.OrderStatusPaid

		err = m.mDB.UpdateOrderInfo(dOrder, core.OrderStatusCreated)
		if err != nil {
			log.Errorf("checkOrderPaid UpdateOrderInfo %s err:%s", dOrder.ID, err.Error())
		}

	}
}

func (m *Mgr) createSpaceFromOrders() {
	orders, err := m.mDB.LoadOrdersByStatus(core.OrderStatusPaid)
	if err != nil {
		log.Errorf("createSpaceFromOrders LoadOrderIDsByStatus err:%s", err.Error())
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		status := core.OrderStatusDone

		err = m.kubMgr.CreateSpaceAndResourceQuotas(order.WorkspaceID, order.Account, order.Cluster, order.CPUCores, order.RAMSize, order.StorageSize)
		if err != nil {
			log.Errorf("createSpaceFromOrders CreateSpaceAndResourceQuotas %s err:%s", order.ID, err.Error())

			status = core.OrderStatusFailed
		}

		err = m.mDB.UpdateOrderStatus(order.ID, status, core.OrderStatusPaid)
		if err != nil {
			log.Errorf("createSpaceFromOrders UpdateOrderStatus %s err:%s", order.ID, err.Error())
		}
	}
}

func (m *Mgr) checkOrderRenewal() {
	ids, err := m.mDB.LoadOrderIDsByStatus(core.OrderStatusRenewal)
	if err != nil {
		log.Errorf("checkOrderRenewal LoadOrderIDsByStatus err:%s", err.Error())
		return
	}

	if len(ids) == 0 {
		return
	}

	tList, err := m.chainMgr.GetOrders(ids)
	if err != nil {
		return
	}

	for _, tOrder := range tList {
		dOrder, err := m.mDB.LoadOrderByID(tOrder.ID)
		if err != nil {
			log.Errorf("checkOrderRenewal LoadOrderByID err:%s", err.Error())
			continue
		}

		if tOrder.Status == chain.Expired {
			m.Terminate(dOrder.ID, dOrder.WorkspaceID, dOrder.Cluster, core.OrderStatusExpired, dOrder.Status)
			continue
		}

		// 超过 x 时间没处理, 表示没付款续期
		duration := time.Now().Sub(dOrder.UpdatedAt)
		if duration.Minutes() > 10 {
			err = m.mDB.UpdateOrderStatus(dOrder.ID, core.OrderStatusDone, core.OrderStatusRenewal)
			if err != nil {
				log.Errorf("checkOrderRenewal UpdateOrderStatus %s err:%s", tOrder.ID, err.Error())
			}
			continue
		}

		hourDuration := int(tOrder.Duration) / 600
		if hourDuration <= dOrder.Duration {
			continue
		}

		dOrder.Status = core.OrderStatusDone
		dOrder.Duration = hourDuration
		dOrder.ExpiredAt = dOrder.CreatedAt.Add(time.Hour * time.Duration(dOrder.Duration))
		dOrder.Price = int(tOrder.LockedFunds)

		err = m.mDB.UpdateOrderInfo(dOrder, core.OrderStatusRenewal)
		if err != nil {
			log.Errorf("checkOrderRenewal UpdateOrderInfo %s err:%s", dOrder.ID, err.Error())
		}
	}
}

func (m *Mgr) checkOrderAbandoned() {
	ids, err := m.mDB.LoadOrderIDsByStatus(core.OrderStatusAbandoned)
	if err != nil {
		log.Errorf("checkOrderAbandoned LoadOrderIDsByStatus err:%s", err.Error())
		return
	}

	if len(ids) == 0 {
		return
	}

	tList, err := m.chainMgr.GetOrders(ids)
	if err != nil {
		return
	}

	for _, tOrder := range tList {
		log.Infof("checkOrderAbandoned %s , status %s", tOrder.ID, tOrder.Status)
		// 升级成功, 旧订单已过期
		if tOrder.Status == chain.Expired {
			err = m.mDB.UpdateOrderStatus(tOrder.ID, core.OrderStatusExpired, core.OrderStatusAbandoned)
			if err != nil {
				log.Errorf("checkOrderAbandoned UpdateOrderStatus %s err:%s", tOrder.ID, err.Error())
			}

			continue
		}

		dOrder, err := m.mDB.LoadOrderByID(tOrder.ID)
		if err != nil {
			log.Errorf("checkOrderAbandoned LoadOrderByID %s err:%s", tOrder.ID, err.Error())
			continue
		}

		// 超过 x 时间没处理, 表示没付款升级
		duration := time.Now().Sub(dOrder.UpdatedAt)
		if duration.Minutes() > 10 {
			err = m.mDB.UpdateOrderStatus(dOrder.ID, core.OrderStatusDone, core.OrderStatusAbandoned)
			if err != nil {
				log.Errorf("checkOrderAbandoned UpdateOrderStatus %s err:%s", tOrder.ID, err.Error())
			}
		}

		// mDB.UpdateOrderUpdated(tOrder.ID)
	}
}

func (m *Mgr) checkOrderUpgrade() {
	ids, err := m.mDB.LoadOrderIDsByStatus(core.OrderStatusUpgrade)
	if err != nil {
		log.Errorf("checkOrderUpgrade LoadOrderIDsByStatus err:%s", err.Error())
		return
	}

	if len(ids) == 0 {
		return
	}

	tList, err := m.chainMgr.GetOrders(ids)
	if err != nil {
		return
	}

	for _, tOrder := range tList {
		dOrder, err := m.mDB.LoadOrderByID(tOrder.ID)
		if err != nil {
			log.Errorf("checkOrderUpgrade LoadOrderByID %s err:%s", tOrder.ID, err.Error())
			continue
		}

		if tOrder.Status == chain.Expired {
			err = m.mDB.UpdateOrderStatus(tOrder.ID, core.OrderStatusExpired, core.OrderStatusUpgrade)
			if err != nil {
				log.Errorf("checkOrderUpgrade UpdateOrderStatus %s err:%s", tOrder.ID, err.Error())
			}

			continue
		}

		hourDuration := int(tOrder.Duration) / 600

		dOrder.Status = core.OrderStatusDone
		dOrder.Duration = hourDuration
		dOrder.ExpiredAt = dOrder.CreatedAt.Add(time.Hour * time.Duration(tOrder.Duration))
		dOrder.Price = int(tOrder.LockedFunds)
		dOrder.CPUCores = int(tOrder.Resource.CPU)
		dOrder.StorageSize = int(tOrder.Resource.Disk)
		dOrder.RAMSize = int(tOrder.Resource.Memory)

		err = m.kubMgr.UpdateUserResourceQuotas(dOrder.WorkspaceID, dOrder.Cluster, dOrder.CPUCores, dOrder.RAMSize, dOrder.StorageSize)
		if err != nil {
			log.Errorf("checkOrderUpgrade UpdateUserResourceQuotas %s err:%s", tOrder.ID, err.Error())
			continue
		}

		err = m.mDB.UpdateOrderInfo(dOrder, core.OrderStatusUpgrade)
		if err != nil {
			log.Errorf("checkOrderUpgrade UpdateOrderInfo %s err:%s", tOrder.ID, err.Error())
		}
	}
}

func (m *Mgr) checkOrderActive() {
	ids, err := m.mDB.LoadOrderIDsByStatus(core.OrderStatusDone)
	if err != nil {
		log.Errorf("checkOrderActive LoadOrderIDsByStatus err:%s", err.Error())
		return
	}

	if len(ids) == 0 {
		return
	}

	list, err := m.chainMgr.GetOrders(ids)
	if err != nil {
		return
	}

	for _, info := range list {
		if info.Status == chain.Active {
			m.mDB.UpdateOrderUpdated(info.ID)
			continue
		}

		order, err := m.mDB.LoadOrderByID(info.ID)
		if err != nil {
			log.Errorf("checkOrderActive LoadOrderByID %s err:%s", order.ID, err.Error())
			continue
		}

		err = m.Terminate(info.ID, order.WorkspaceID, order.Cluster, core.OrderStatusExpired, core.OrderStatusDone)
		if err != nil {
			log.Errorf("checkOrderActive Terminate %s err:%s", info.ID, err.Error())
		}

	}
}

func (m *Mgr) checkOrderExpired() {
	orders, err := m.mDB.LoadExpiredOrders([]core.OrderStatus{core.OrderStatusDone, core.OrderStatusFailed, core.OrderStatusUpgrade, core.OrderStatusRenewal})
	if err != nil {
		log.Errorf("checkOrderExpired LoadExpiredOrders err:%s", err.Error())
		return
	}

	if len(orders) == 0 {
		return
	}

	for _, order := range orders {
		m.Terminate(order.ID, order.WorkspaceID, order.Cluster, core.OrderStatusExpired, order.Status)
	}
}

func (m *Mgr) deleteOrders() {
	err := m.mDB.DeleteOrdersByCreated(time.Now().Add(-5 * time.Minute))
	if err != nil {
		log.Errorf("deleteOrders DeleteOrdersByCreated err:%s", err.Error())
	}
}

// func (m *Mgr) cleanOrders() {
// 	err := mDB.CleanOrders()
// 	if err != nil {
// 		log.Errorf("cleanOrders CleanOrders err:%s", err.Error())
// 	}
// }
