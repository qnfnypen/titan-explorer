package order

import (
	"context"
	"math"
	"time"

	"github.com/gnasnik/titan-explorer/core"
)

// Terminate deletes a user space by ID.
func (m *Mgr) Terminate(orderID, workspaceID, cluster string, status core.OrderStatus, oldStatus core.OrderStatus) error {
	err := m.kubMgr.DeleteUserSpace(workspaceID, cluster)
	if err != nil {
		log.Errorf("Terminate DeleteUserSpace %s err:%s", orderID, err.Error())
	}

	err = m.chainMgr.ReleaseOrder(orderID)
	if err != nil {
		log.Errorf("Terminate ReleaseOrder %s err:%s", orderID, err.Error())
	}

	err = m.mDB.UpdateOrderStatus(orderID, status, oldStatus)
	if err != nil {
		log.Errorf("Terminate UpdateOrderStatus %s err:%s", orderID, err.Error())
	}

	return err
}

// Renewal renews an order identified by the given ID.
func (m *Mgr) Renewal(id string) error {
	return m.mDB.UpdateOrderStatus(id, core.OrderStatusRenewal, core.OrderStatusDone)
}

// Upgrade creates an upgraded order with the given parameters.
func (m *Mgr) Upgrade(oldOrder *core.Order, newOrder *core.OrderInfoReq, account, orderID string, price int) error {
	err := m.createOrder(newOrder, account, orderID, oldOrder.WorkspaceID, oldOrder.Cluster, price, core.OrderStatusUpgrade)
	if err != nil {
		return err
	}

	return m.mDB.UpdateOrderStatus(oldOrder.ID, core.OrderStatusAbandoned, core.OrderStatusDone)
}

// Create creates a new order with the given parameters.
func (m *Mgr) Create(params *core.OrderInfoReq, account, orderID string, price int) error {
	return m.createOrder(params, account, orderID, orderID, m.kubMgr.GetCluster(), price, core.OrderStatusCreated)
}

func (m *Mgr) createOrder(params *core.OrderInfoReq, account, orderID, workspaceID, cluster string, price int, status core.OrderStatus) error {
	order := &core.Order{
		Account:     account,
		CPUCores:    params.CPUCores,
		RAMSize:     params.RAMSize,
		StorageSize: params.StorageSize,
		Duration:    params.Duration,
		Status:      status,
		ID:          orderID,
		Price:       price,
		Cluster:     cluster,
		WorkspaceID: workspaceID,
		ExpiredAt:   time.Now().Add(time.Duration(params.Duration) * time.Hour),
	}

	return m.mDB.CreateOrder(context.Background(), order)
}

// CalculateOrderRefund calculates the refund amount for a given order based on its creation time.
func (m *Mgr) CalculateOrderRefund(order *core.Order) int {
	duration := time.Now().Sub(order.CreatedAt)
	hours := duration.Hours()
	hour := int(math.Ceil(hours))

	pr := m.CalculateTotalCost(&core.OrderInfoReq{
		CPUCores:    order.CPUCores,
		RAMSize:     order.RAMSize,
		StorageSize: order.StorageSize,
		Duration:    hour,
	})

	refund := order.Price - pr
	if refund < 0 {
		refund = 0
	}

	return refund
}
