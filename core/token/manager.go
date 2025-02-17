package token

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/gnasnik/titan-explorer/core/errors"

	"github.com/gnasnik/titan-explorer/core"
	"github.com/gnasnik/titan-explorer/core/chain"
	"github.com/gnasnik/titan-explorer/core/dao"

	logging "github.com/ipfs/go-log/v2"
)

var log = logging.Logger("token")

const (
	hourlyQuota  = 10000
	maxUserQuota = 400
)

// Mgr manages token resources.
type Mgr struct {
	mDB      *dao.Mgr
	chainMgr *chain.Mgr
}

// NewTokenManager creates a new instance of Mgr for managing tokens.
func NewTokenManager(db *dao.Mgr, c *chain.Mgr) *Mgr {
	m := &Mgr{}

	m.mDB = db
	m.chainMgr = c

	return m
}

func (m *Mgr) getCurrentHourString() string {
	now := time.Now()
	// time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())

	return fmt.Sprintf("%d-%d-%d-%d", now.Year(), now.Month(), now.Day(), now.Hour())
}

func (m *Mgr) getNextHour() time.Time {
	nextHourTime := time.Now().Add(time.Hour)
	year, month, day := nextHourTime.Date()
	hour, _, _ := nextHourTime.Clock()

	return time.Date(year, month, day, hour, 0, 0, 0, time.Local)
}

// GetAmountDistributedInfo calculates the distributed amount for the current hour and returns the distributed amount for the current hour.
func (m *Mgr) GetAmountDistributedInfo(account string) (*core.AmountDistributedInfo, error) {
	currentHour := m.getCurrentHourString()

	amount, err := m.mDB.GetOrCreateHourlyQuota(currentHour)
	if err != nil {
		return nil, err
	}

	nextTime := m.getNextHour()

	received := m.isReceive(account)

	return &core.AmountDistributedInfo{RemainingAmount: hourlyQuota - amount, UsedAmount: amount, NextTime: nextTime, Received: received}, nil
}

func (m *Mgr) isReceive(account string) bool {
	lastReceive, err := m.mDB.GetLastReceiveByAccount(account)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("isReceive %s err:%s", account, err.Error())
		return true
	}

	if !lastReceive.IsZero() {
		if m.isToday(lastReceive) {
			return true
		}
	}

	return false
}

// ReceiveTokens allows the specified account to receive tokens and returns the number of tokens received.
func (m *Mgr) ReceiveTokens(account string) (int, error) {
	currentHour := m.getCurrentHourString()

	distributedAmount, err := m.mDB.GetOrCreateHourlyQuota(currentHour)
	if err != nil {
		return errors.InternalServer, err
	}

	if distributedAmount >= hourlyQuota {
		return errors.QuotaIssued, nil
	}

	lastReceive, err := m.mDB.GetLastReceiveByAccount(account)
	if err != nil && err != sql.ErrNoRows {
		log.Errorf("isReceive %s err:%s", account, err.Error())
		return errors.InternalServer, err
	}

	if !lastReceive.IsZero() {
		if m.isToday(lastReceive) {
			return errors.Received, nil
		}
	}

	if distributedAmount+maxUserQuota > hourlyQuota {
		return errors.QuotaIssued, nil
	}

	if err == sql.ErrNoRows {
		err = m.mDB.SaveAccountReceive(account, maxUserQuota)
	} else {
		err = m.mDB.UpdateAccountReceive(account, maxUserQuota)
	}
	if err != nil {
		return errors.InternalServer, err
	}

	err = m.mDB.UpdateHourlyQuota(maxUserQuota, currentHour)
	if err != nil {
		return errors.InternalServer, err
	}

	hash, err := m.chainMgr.ReceiveTokens(account, strconv.Itoa(maxUserQuota))
	if err != nil {
		return errors.InternalServer, err
	}

	err = m.mDB.SaveReceiveHistory(&core.ReceiveHistory{Account: account, Amount: maxUserQuota, Hash: hash})
	if err != nil {
		return errors.InternalServer, err
	}

	return errors.Success, nil
}

func (m *Mgr) isToday(t time.Time) bool {
	now := time.Now()
	year, month, day := now.Date()
	today := time.Date(year, month, day, 0, 0, 0, 0, time.Local)
	tomorrow := today.AddDate(0, 0, 1)
	return t.After(today) && t.Before(tomorrow)
}

// GetBalance retrieves the balance for a given account.
func (m *Mgr) GetBalance(account string) (string, error) {
	return m.chainMgr.GetBalance(account)
}
