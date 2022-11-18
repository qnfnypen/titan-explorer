// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"
	"database/sql"

	"gorm.io/gorm"

	"gorm.io/gen"

	"gorm.io/plugin/dbresolver"
)

var (
	Q             = new(Query)
	DeviceInfo    *deviceInfo
	HourDaily     *hourDaily
	IncomeDaily   *incomeDaily
	LoginLog      *loginLog
	OperationLog  *operationLog
	RetrievalInfo *retrievalInfo
	Scheduler     *scheduler
	TaskInfo      *taskInfo
	User          *user
)

func SetDefault(db *gorm.DB, opts ...gen.DOOption) {
	*Q = *Use(db, opts...)
	DeviceInfo = &Q.DeviceInfo
	HourDaily = &Q.HourDaily
	IncomeDaily = &Q.IncomeDaily
	LoginLog = &Q.LoginLog
	OperationLog = &Q.OperationLog
	RetrievalInfo = &Q.RetrievalInfo
	Scheduler = &Q.Scheduler
	TaskInfo = &Q.TaskInfo
	User = &Q.User
}

func Use(db *gorm.DB, opts ...gen.DOOption) *Query {
	return &Query{
		db:            db,
		DeviceInfo:    newDeviceInfo(db, opts...),
		HourDaily:     newHourDaily(db, opts...),
		IncomeDaily:   newIncomeDaily(db, opts...),
		LoginLog:      newLoginLog(db, opts...),
		OperationLog:  newOperationLog(db, opts...),
		RetrievalInfo: newRetrievalInfo(db, opts...),
		Scheduler:     newScheduler(db, opts...),
		TaskInfo:      newTaskInfo(db, opts...),
		User:          newUser(db, opts...),
	}
}

type Query struct {
	db *gorm.DB

	DeviceInfo    deviceInfo
	HourDaily     hourDaily
	IncomeDaily   incomeDaily
	LoginLog      loginLog
	OperationLog  operationLog
	RetrievalInfo retrievalInfo
	Scheduler     scheduler
	TaskInfo      taskInfo
	User          user
}

func (q *Query) Available() bool { return q.db != nil }

func (q *Query) clone(db *gorm.DB) *Query {
	return &Query{
		db:            db,
		DeviceInfo:    q.DeviceInfo.clone(db),
		HourDaily:     q.HourDaily.clone(db),
		IncomeDaily:   q.IncomeDaily.clone(db),
		LoginLog:      q.LoginLog.clone(db),
		OperationLog:  q.OperationLog.clone(db),
		RetrievalInfo: q.RetrievalInfo.clone(db),
		Scheduler:     q.Scheduler.clone(db),
		TaskInfo:      q.TaskInfo.clone(db),
		User:          q.User.clone(db),
	}
}

func (q *Query) ReadDB() *Query {
	return q.clone(q.db.Clauses(dbresolver.Read))
}

func (q *Query) WriteDB() *Query {
	return q.clone(q.db.Clauses(dbresolver.Write))
}

func (q *Query) ReplaceDB(db *gorm.DB) *Query {
	return &Query{
		db:            db,
		DeviceInfo:    q.DeviceInfo.replaceDB(db),
		HourDaily:     q.HourDaily.replaceDB(db),
		IncomeDaily:   q.IncomeDaily.replaceDB(db),
		LoginLog:      q.LoginLog.replaceDB(db),
		OperationLog:  q.OperationLog.replaceDB(db),
		RetrievalInfo: q.RetrievalInfo.replaceDB(db),
		Scheduler:     q.Scheduler.replaceDB(db),
		TaskInfo:      q.TaskInfo.replaceDB(db),
		User:          q.User.replaceDB(db),
	}
}

type queryCtx struct {
	DeviceInfo    IDeviceInfoDo
	HourDaily     IHourDailyDo
	IncomeDaily   IIncomeDailyDo
	LoginLog      ILoginLogDo
	OperationLog  IOperationLogDo
	RetrievalInfo IRetrievalInfoDo
	Scheduler     ISchedulerDo
	TaskInfo      ITaskInfoDo
	User          IUserDo
}

func (q *Query) WithContext(ctx context.Context) *queryCtx {
	return &queryCtx{
		DeviceInfo:    q.DeviceInfo.WithContext(ctx),
		HourDaily:     q.HourDaily.WithContext(ctx),
		IncomeDaily:   q.IncomeDaily.WithContext(ctx),
		LoginLog:      q.LoginLog.WithContext(ctx),
		OperationLog:  q.OperationLog.WithContext(ctx),
		RetrievalInfo: q.RetrievalInfo.WithContext(ctx),
		Scheduler:     q.Scheduler.WithContext(ctx),
		TaskInfo:      q.TaskInfo.WithContext(ctx),
		User:          q.User.WithContext(ctx),
	}
}

func (q *Query) Transaction(fc func(tx *Query) error, opts ...*sql.TxOptions) error {
	return q.db.Transaction(func(tx *gorm.DB) error { return fc(q.clone(tx)) }, opts...)
}

func (q *Query) Begin(opts ...*sql.TxOptions) *QueryTx {
	return &QueryTx{q.clone(q.db.Begin(opts...))}
}

type QueryTx struct{ *Query }

func (q *QueryTx) Commit() error {
	return q.db.Commit().Error
}

func (q *QueryTx) Rollback() error {
	return q.db.Rollback().Error
}

func (q *QueryTx) SavePoint(name string) error {
	return q.db.SavePoint(name).Error
}

func (q *QueryTx) RollbackTo(name string) error {
	return q.db.RollbackTo(name).Error
}
