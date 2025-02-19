package core

import "time"

// UserNonce represents a user's account and its associated nonce.
type UserNonce struct {
	Account   string    `db:"account" `
	Nonce     string    `db:"nonce"`
	ExpiredAt time.Time `db:"expired_at" json:"expired_at"`
}

// User represents a user in the system.
type User struct {
	Account     string    `db:"account" json:"account"`
	Avatar      string    `db:"avatar" json:"avatar"`
	Username    string    `db:"user_name" json:"user_name"`
	UserEmail   string    `db:"user_email" json:"user_email"`
	KubPwd      string    `db:"kub_pwd" json:"kub_pwd"`
	StorageUser string    `db:"storage_user" json:"storage_user"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

// OrderInfoReq represents a request to create or get price an order with specified resources.
type OrderInfoReq struct {
	CPUCores    int `json:"cpu"`
	RAMSize     int `json:"ram"`      // in GB
	StorageSize int `json:"storage"`  // in GB
	Duration    int `json:"duration"` // Hour
}

// UpgradeOrderInfoReq represents a request for upgrading order information.
type UpgradeOrderInfoReq struct {
	CPUCores    int `json:"cpu"`
	RAMSize     int `json:"ram"`      // in GB
	StorageSize int `json:"storage"`  // in GB
	Duration    int `json:"duration"` // Hour

	ID string `json:"id"`
}

// OrderIDReq represents a request for an order ID.
type OrderIDReq struct {
	ID string `json:"id"`
}

// OrderHashReq represents a request for an order hash.
type OrderHashReq struct {
	Hash string `json:"hash"`
	ID   string `json:"id"`
}

// Order represents a customer's order in the system.
type Order struct {
	ID          string      `db:"id" json:"id"`
	Account     string      `db:"account" json:"account"`
	CPUCores    int         `db:"cpu" json:"cpu"`
	RAMSize     int         `db:"ram" json:"ram"`
	StorageSize int         `db:"storage" json:"storage"`
	Duration    int         `db:"duration" json:"duration"` // Hour
	Price       int         `db:"price" json:"price"`
	Cluster     string      `db:"cluster" json:"cluster"`
	Status      OrderStatus `db:"status" json:"status"`
	WorkspaceID string      `db:"workspace_id" json:"workspace_id"`
	Hash        string      `db:"hash" json:"hash"`
	CreatedAt   time.Time   `db:"created_at" json:"created_at"`
	ExpiredAt   time.Time   `db:"expired_at" json:"expired_at"`
	UpdatedAt   time.Time   `db:"updated_at" json:"updated_at"`
}

// OrderStatus represents the status of an order.
type OrderStatus int

const (
	// OrderStatusCreated indicates that the order has been created.
	OrderStatusCreated OrderStatus = iota
	// OrderStatusPaid indicates that the order has been paid.
	OrderStatusPaid
	// OrderStatusDone indicates that the order has been completed. (Active)
	OrderStatusDone
	// OrderStatusExpired indicates that the order has expired.
	OrderStatusExpired
	// OrderStatusFailed indicates that the order has creation failed.
	OrderStatusFailed
	// OrderStatusTimeout indicates that the order has payment timeout.
	OrderStatusTimeout
	// OrderStatusUpgrade indicates that the order has been upgrade. (Active)
	OrderStatusUpgrade
	// OrderStatusRenewal indicates that the order has been renewal. (Active)
	OrderStatusRenewal
	// OrderStatusAbandoned indicates that the order has been abandoned. (Active)
	OrderStatusAbandoned
	// OrderStatusTermination indicates that the order has early termination.
	OrderStatusTermination
)

// ReceiveHistory represents the history of receive for an account.
type ReceiveHistory struct {
	Account   string    `db:"account" json:"account"`
	Amount    int       `db:"amount" json:"amount"`
	Hash      string    `db:"hash" json:"hash"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

// AmountDistributedInfo holds information about the distribution of amounts.
type AmountDistributedInfo struct {
	RemainingAmount int       `json:"remaining_amount"`
	UsedAmount      int       `json:"used_amount"`
	NextTime        time.Time `json:"next_time"`
	Received        bool      `json:"received"`
}
