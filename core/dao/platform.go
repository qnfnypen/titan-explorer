package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gnasnik/titan-explorer/config"
	"github.com/gnasnik/titan-explorer/core"
	"github.com/jmoiron/sqlx"
)

const (
	userInfoTable       = "container_platform_users"
	orderInfoTable      = "container_platform_orders"
	userReceiveTable    = "container_platform_user_receive"
	hourlyQuotasTable   = "container_platform_hourly_quotas"
	receiveHistoryTable = "container_platform_receive_history"
	// userNonceTable      = "user_nonce"
)

// Mgr manages database operations.
type Mgr struct {
	db *sqlx.DB
}

// NewDbMgr creates a new db instance.
func NewDbMgr(cfg *config.Config) (*Mgr, error) {
	n := new(Mgr)
	pdb, err := sqlx.Connect("mysql", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	setDBConfig(pdb, maxOpenConnections, connMaxLifetime, maxIdleConnections, connMaxIdleTime)
	n.db = pdb

	go n.startCheckNodeTimer()

	return n, err
}

// CleanData performs a cleanup of outdated records across various tables based on predefined intervals.
func (n *Mgr) cleanData() {
	query := fmt.Sprintf(`DELETE FROM %s WHERE created_at<DATE_SUB(NOW(), INTERVAL 2 DAY) `, hourlyQuotasTable)
	_, err := n.db.Exec(query)
	if err != nil {
		log.Warnf("cleanData hourlyQuotasTable err:%s", err.Error())
	}

	query = fmt.Sprintf(`DELETE FROM %s WHERE created_at<DATE_SUB(NOW(), INTERVAL 30 DAY) `, receiveHistoryTable)
	_, err = n.db.Exec(query)
	if err != nil {
		log.Warnf("cleanData receiveHistoryTable err:%s", err.Error())
	}

	// query = fmt.Sprintf(`DELETE FROM %s WHERE expired_at<DATE_SUB(NOW(), INTERVAL 1 DAY) `, userNonceTable)
	// _, err = n.db.Exec(query)
	// if err != nil {
	// 	log.Warnf("cleanData userNonceTable err:%s", err.Error())
	// }

	query = fmt.Sprintf(`DELETE FROM %s WHERE last_receive<DATE_SUB(NOW(), INTERVAL 5 DAY) `, userReceiveTable)
	_, err = n.db.Exec(query)
	if err != nil {
		log.Warnf("cleanData userReceiveTable err:%s", err.Error())
	}

	query = fmt.Sprintf(`DELETE FROM %s WHERE (status=? or status=?) AND updated_at<DATE_SUB(NOW(), INTERVAL 20 DAY) `, orderInfoTable)
	_, err = n.db.Exec(query, core.OrderStatusExpired, core.OrderStatusTermination)
	if err != nil {
		log.Warnf("CleanData orderInfoTable err:%s", err.Error())
	}
}

func (n *Mgr) startCheckNodeTimer() {
	now := time.Now()

	nextTime := time.Date(now.Year(), now.Month(), now.Day(), 17, 30, 0, 0, now.Location())
	if now.After(nextTime) {
		nextTime = nextTime.Add(24 * time.Hour)
	}

	duration := nextTime.Sub(now)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	for {
		<-timer.C

		n.cleanData()

		timer.Reset(24 * time.Hour)
	}
}

// CreateOrder creates a new order in the database.
func (n *Mgr) CreateOrder(ctx context.Context, order *core.Order) error {
	query := fmt.Sprintf(`INSERT INTO %s (id, account, cpu, ram, storage, duration, status, price, expired_at, cluster, workspace_id)
			VALUES (:id, :account, :cpu, :ram, :storage, :duration, :status, :price, :expired_at, :cluster, :workspace_id);`, orderInfoTable)
	_, err := n.db.NamedExec(query, order)

	return err
}

// UpdateOrderUpdated updates the updated of an order in the database.
func (n *Mgr) UpdateOrderUpdated(id string) error {
	query := fmt.Sprintf(`UPDATE %s SET updated_at=NOW() WHERE id=?`, orderInfoTable)
	_, err := n.db.Exec(query, id)

	return err
}

// UpdateOrderInfo updates the info of an order in the database.
func (n *Mgr) UpdateOrderInfo(order *core.Order, oldStatus core.OrderStatus) error {
	query := fmt.Sprintf(`UPDATE %s SET status=?,cpu=?,storage=?,duration=?,price=?,expired_at=?,updated_at=NOW() WHERE id=? AND status=? `, orderInfoTable)
	_, err := n.db.Exec(query, order.Status, order.CPUCores, order.StorageSize, order.Duration, order.Price, order.ExpiredAt, order.ID, oldStatus)

	return err
}

// LoadExpiredOrders retrieves a list of expired order IDs.
func (n *Mgr) LoadExpiredOrders() ([]*core.Order, error) {
	var infos []*core.Order

	query := fmt.Sprintf("SELECT * FROM %s WHERE status=? AND expired_at<NOW()", orderInfoTable)
	err := n.db.Select(&infos, query, core.OrderStatusDone)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// LoadOrderIDsByStatus retrieves order ids based on their status.
func (n *Mgr) LoadOrderIDsByStatus(status core.OrderStatus) ([]string, error) {
	var ids []string

	query := fmt.Sprintf("SELECT id FROM %s WHERE status=?", orderInfoTable)
	err := n.db.Select(&ids, query, status)
	if err != nil {
		return nil, err
	}

	return ids, nil
}

// LoadOrdersByStatus retrieves orders based on their status.
func (n *Mgr) LoadOrdersByStatus(status core.OrderStatus) ([]*core.Order, error) {
	var infos []*core.Order

	query := fmt.Sprintf("SELECT * FROM %s WHERE status=?", orderInfoTable)
	err := n.db.Select(&infos, query, status)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

// LoadOrderByID retrieves orders based on  id.
func (n *Mgr) LoadOrderByID(id string) (*core.Order, error) {
	var info core.Order

	query := fmt.Sprintf("SELECT * FROM %s WHERE id=?", orderInfoTable)
	err := n.db.Get(&info, query, id)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// UpdateOrderStatus updates the status of an order in the database.
func (n *Mgr) UpdateOrderStatus(id string, status, oldStatus core.OrderStatus) error {
	query := fmt.Sprintf(`UPDATE %s SET status=?, updated_at=NOW() WHERE id=? AND status=? `, orderInfoTable)
	_, err := n.db.Exec(query, status, id, oldStatus)

	return err
}

// UpdateOrderHash updates the hash of an order in the database.
func (n *Mgr) UpdateOrderHash(id, hash string) error {
	query := fmt.Sprintf(`UPDATE %s SET hash=? WHERE id=? `, orderInfoTable)
	_, err := n.db.Exec(query, hash, id)

	return err
}

// CleanOrders deletes orders older than 1 day from the database.
func (n *Mgr) CleanOrders() error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE status=? AND created_at<DATE_SUB(NOW(), INTERVAL 1 DAY) `, orderInfoTable)
	_, err := n.db.Exec(query, core.OrderStatusCreated)

	return err
}

// DeleteOrdersByCreated deletes orders with a specific status created before the given time.
func (n *Mgr) DeleteOrdersByCreated(time time.Time) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE status=? AND created_at<? `, orderInfoTable)
	_, err := n.db.Exec(query, core.OrderStatusCreated, time)

	return err
}

// LoadAccountOrdersByStatuses retrieves account orders by their statuses, with pagination.
func (n *Mgr) LoadAccountOrdersByStatuses(account string, statuses []core.OrderStatus, page, size int) ([]*core.Order, int64, error) {
	infos := make([]*core.Order, 0)

	if page < 1 {
		page = 1
	}

	offset := uint64((page - 1) * size)

	query := fmt.Sprintf("SELECT * FROM %s WHERE account=? AND status in (?) ORDER BY created_at DESC LIMIT ? OFFSET ?", orderInfoTable)
	srQuery, args, err := sqlx.In(query, account, statuses, size, offset)
	if err != nil {
		return nil, 0, err
	}

	srQuery = n.db.Rebind(srQuery)
	err = n.db.Select(&infos, srQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE account=? AND status in (?)", orderInfoTable)
	srQuery, args, err = sqlx.In(countQuery, account, statuses)
	if err != nil {
		return nil, 0, err
	}

	var count int64
	srQuery = n.db.Rebind(srQuery)
	err = n.db.Get(&count, srQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return infos, count, nil
}

// LoadAccountOrdersByStatus retrieves a list of orders for a given account, with pagination.
func (n *Mgr) LoadAccountOrdersByStatus(ctx context.Context, account string, status core.OrderStatus, page, size int) ([]*core.Order, int64, error) {
	out := make([]*core.Order, 0)

	var count int64
	if page < 1 {
		page = 1
	}

	query, args, err := squirrel.Select("*").From(orderInfoTable).Where(squirrel.Eq{"account": account}).Where(squirrel.Eq{"status": status}).Offset(uint64((page - 1) * size)).Limit(uint64(size)).ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := n.db.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}

	sq2 := squirrel.Select("COUNT(*)").From(orderInfoTable).Where(squirrel.Eq{"account": account}).Where(squirrel.Eq{"status": status})

	query2, args2, err := sq2.ToSql()
	if err != nil {
		return nil, 0, err
	}

	err = n.db.Get(&count, query2, args2...)
	if err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

// LoadAccountOrders retrieves a list of orders for a given account, with pagination.
func (n *Mgr) LoadAccountOrders(ctx context.Context, account string, page, size int) ([]*core.Order, int64, error) {
	out := make([]*core.Order, 0)

	var count int64
	if page < 1 {
		page = 1
	}

	query, args, err := squirrel.Select("*").From(orderInfoTable).Where(squirrel.Eq{"account": account}).Offset(uint64((page - 1) * size)).Limit(uint64(size)).ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := n.db.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}

	sq2 := squirrel.Select("COUNT(*)").From(orderInfoTable).Where(squirrel.Eq{"account": account})

	query2, args2, err := sq2.ToSql()
	if err != nil {
		return nil, 0, err
	}

	err = n.db.Get(&count, query2, args2...)
	if err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

// GetOrCreateHourlyQuota retrieves the distributed amount for the current hour or creates a new entry with a zero amount if it doesn't exist.
func (n *Mgr) GetOrCreateHourlyQuota(currentHour string) (int, error) {
	queryS := fmt.Sprintf(`SELECT amount FROM %s WHERE hour = ? `, hourlyQuotasTable)
	queryI := fmt.Sprintf(`INSERT INTO %s (hour, amount) VALUES (?, 0) `, hourlyQuotasTable)

	var distributedAmount int
	err := n.db.QueryRow(queryS, currentHour).Scan(&distributedAmount)
	if err == sql.ErrNoRows {
		_, err = n.db.Exec(queryI, currentHour)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return distributedAmount, nil
}

// UpdateHourlyQuota updates the distributed amount for the given hour.
func (n *Mgr) UpdateHourlyQuota(amount int, currentHour string) error {
	query := fmt.Sprintf(`UPDATE %s SET amount = amount + ? WHERE hour = ? `, hourlyQuotasTable)
	_, err := n.db.Exec(query, amount, currentHour)
	return err
}

// GetLastReceiveByAccount retrieves the last receive time for a given account.
func (n *Mgr) GetLastReceiveByAccount(account string) (time.Time, error) {
	query := fmt.Sprintf(`SELECT last_receive FROM %s WHERE account = ? `, userReceiveTable)
	var lastReceive time.Time
	err := n.db.QueryRow(query, account).Scan(&lastReceive)
	if err != nil && err != sql.ErrNoRows {
		return lastReceive, err
	}

	return lastReceive, err
}

// SaveAccountReceive inserts account receive into the database.
func (n *Mgr) SaveAccountReceive(account string, maxUserQuota int) error {
	query := fmt.Sprintf(`INSERT INTO %s (account, amount, last_receive) VALUES (?, ?, ?) `, userReceiveTable)
	_, err := n.db.Exec(query, account, maxUserQuota, time.Now())
	return err
}

// UpdateAccountReceive updates the receive for a given account by adding to the amount and updating the last Receive time.
func (n *Mgr) UpdateAccountReceive(account string, maxUserQuota int) error {
	query := fmt.Sprintf(`UPDATE %s SET amount = amount + ?, last_receive = ? WHERE account = ? `, userReceiveTable)
	_, err := n.db.Exec(query, maxUserQuota, time.Now(), account)
	return err
}

// SaveReceiveHistory saves the receive history for a given account.
func (n *Mgr) SaveReceiveHistory(info *core.ReceiveHistory) error {
	qry := fmt.Sprintf(`INSERT INTO %s (account, amount, hash) 
		        VALUES (:account, :amount, :hash)`, receiveHistoryTable)
	_, err := n.db.NamedExec(qry, info)

	return err
}

// LoadReceiveHistory retrieves the receive history for a given account, paginated by page and size.
func (n *Mgr) LoadReceiveHistory(ctx context.Context, account string, page, size int) ([]*core.ReceiveHistory, int64, error) {
	out := make([]*core.ReceiveHistory, 0)

	var count int64
	if page < 1 {
		page = 1
	}

	query, args, err := squirrel.Select("*").From(receiveHistoryTable).Where(squirrel.Eq{"account": account}).OrderBy("created_at DESC").Offset(uint64((page - 1) * size)).Limit(uint64(size)).ToSql()
	if err != nil {
		return nil, 0, err
	}

	if err := n.db.SelectContext(ctx, &out, query, args...); err != nil {
		return nil, 0, err
	}

	sq2 := squirrel.Select("COUNT(*)").From(receiveHistoryTable).Where(squirrel.Eq{"account": account})

	query2, args2, err := sq2.ToSql()
	if err != nil {
		return nil, 0, err
	}

	err = n.db.Get(&count, query2, args2...)
	if err != nil {
		return nil, 0, err
	}

	return out, count, nil
}

// GetUserInfo retrieves user information based on the account provided.
func (n *Mgr) GetUserInfo(ctx context.Context, account string) (*core.User, error) {
	response := core.User{}

	if err := n.db.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE account = ?`, userInfoTable), account,
	).StructScan(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// CreateUser creates a new user in the database.
func (n *Mgr) CreateUser(ctx context.Context, user *core.User) error {
	query := fmt.Sprintf(`INSERT INTO %s (account, user_name, user_email, avatar, kub_pwd, storage_user) VALUES (:account, :user_name, :user_email, :avatar, :kub_pwd, :storage_user) `, userInfoTable)
	_, err := n.db.NamedExec(query, user)

	return err
}

// GetUserByAccount retrieves a user by their account from the database.
func (n *Mgr) GetUserByAccount(ctx context.Context, account string) (*core.User, error) {
	var out core.User
	if err := n.db.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE account = ?`, userInfoTable), account,
	).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

// GetUserByStorageUser retrieves a user by their storage_user from the database.
func (n *Mgr) GetUserByStorageUser(ctx context.Context, su string) (*core.User, error) {
	var out core.User
	if err := n.db.QueryRowxContext(ctx, fmt.Sprintf(
		`SELECT * FROM %s WHERE storage_user = ?`, userInfoTable), su,
	).StructScan(&out); err != nil {
		return nil, err
	}

	return &out, nil
}

// UpdateKubPwd updates the Kubernetes password for a given account.
func (n *Mgr) UpdateKubPwd(account, kubPwd string) error {
	query := fmt.Sprintf(`UPDATE %s SET kub_pwd=? WHERE account=? `, userInfoTable)
	_, err := n.db.Exec(query, kubPwd, account)

	return err
}
