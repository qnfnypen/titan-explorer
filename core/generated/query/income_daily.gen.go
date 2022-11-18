// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"github.com/gnasnik/titan-explorer/core/generated/model"
)

func newIncomeDaily(db *gorm.DB, opts ...gen.DOOption) incomeDaily {
	_incomeDaily := incomeDaily{}

	_incomeDaily.incomeDailyDo.UseDB(db, opts...)
	_incomeDaily.incomeDailyDo.UseModel(&model.IncomeDaily{})

	tableName := _incomeDaily.incomeDailyDo.TableName()
	_incomeDaily.ALL = field.NewAsterisk(tableName)
	_incomeDaily.ID = field.NewInt64(tableName, "id")
	_incomeDaily.CreatedAt = field.NewTime(tableName, "created_at")
	_incomeDaily.UpdatedAt = field.NewTime(tableName, "updated_at")
	_incomeDaily.DeletedAt = field.NewField(tableName, "deleted_at")
	_incomeDaily.UserID = field.NewString(tableName, "user_id")
	_incomeDaily.DeviceID = field.NewString(tableName, "device_id")
	_incomeDaily.Time = field.NewTime(tableName, "time")
	_incomeDaily.Income = field.NewFloat64(tableName, "income")
	_incomeDaily.OnlineTime = field.NewFloat64(tableName, "online_time")
	_incomeDaily.PkgLossRatio = field.NewFloat64(tableName, "pkg_loss_ratio")
	_incomeDaily.Latency = field.NewFloat64(tableName, "latency")
	_incomeDaily.NatRatio = field.NewFloat64(tableName, "nat_ratio")
	_incomeDaily.DiskUsage = field.NewFloat64(tableName, "disk_usage")

	_incomeDaily.fillFieldMap()

	return _incomeDaily
}

type incomeDaily struct {
	incomeDailyDo

	ALL          field.Asterisk
	ID           field.Int64
	CreatedAt    field.Time
	UpdatedAt    field.Time
	DeletedAt    field.Field
	UserID       field.String
	DeviceID     field.String
	Time         field.Time
	Income       field.Float64
	OnlineTime   field.Float64
	PkgLossRatio field.Float64
	Latency      field.Float64
	NatRatio     field.Float64
	DiskUsage    field.Float64

	fieldMap map[string]field.Expr
}

func (i incomeDaily) Table(newTableName string) *incomeDaily {
	i.incomeDailyDo.UseTable(newTableName)
	return i.updateTableName(newTableName)
}

func (i incomeDaily) As(alias string) *incomeDaily {
	i.incomeDailyDo.DO = *(i.incomeDailyDo.As(alias).(*gen.DO))
	return i.updateTableName(alias)
}

func (i *incomeDaily) updateTableName(table string) *incomeDaily {
	i.ALL = field.NewAsterisk(table)
	i.ID = field.NewInt64(table, "id")
	i.CreatedAt = field.NewTime(table, "created_at")
	i.UpdatedAt = field.NewTime(table, "updated_at")
	i.DeletedAt = field.NewField(table, "deleted_at")
	i.UserID = field.NewString(table, "user_id")
	i.DeviceID = field.NewString(table, "device_id")
	i.Time = field.NewTime(table, "time")
	i.Income = field.NewFloat64(table, "income")
	i.OnlineTime = field.NewFloat64(table, "online_time")
	i.PkgLossRatio = field.NewFloat64(table, "pkg_loss_ratio")
	i.Latency = field.NewFloat64(table, "latency")
	i.NatRatio = field.NewFloat64(table, "nat_ratio")
	i.DiskUsage = field.NewFloat64(table, "disk_usage")

	i.fillFieldMap()

	return i
}

func (i *incomeDaily) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := i.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (i *incomeDaily) fillFieldMap() {
	i.fieldMap = make(map[string]field.Expr, 13)
	i.fieldMap["id"] = i.ID
	i.fieldMap["created_at"] = i.CreatedAt
	i.fieldMap["updated_at"] = i.UpdatedAt
	i.fieldMap["deleted_at"] = i.DeletedAt
	i.fieldMap["user_id"] = i.UserID
	i.fieldMap["device_id"] = i.DeviceID
	i.fieldMap["time"] = i.Time
	i.fieldMap["income"] = i.Income
	i.fieldMap["online_time"] = i.OnlineTime
	i.fieldMap["pkg_loss_ratio"] = i.PkgLossRatio
	i.fieldMap["latency"] = i.Latency
	i.fieldMap["nat_ratio"] = i.NatRatio
	i.fieldMap["disk_usage"] = i.DiskUsage
}

func (i incomeDaily) clone(db *gorm.DB) incomeDaily {
	i.incomeDailyDo.ReplaceConnPool(db.Statement.ConnPool)
	return i
}

func (i incomeDaily) replaceDB(db *gorm.DB) incomeDaily {
	i.incomeDailyDo.ReplaceDB(db)
	return i
}

type incomeDailyDo struct{ gen.DO }

type IIncomeDailyDo interface {
	gen.SubQuery
	Debug() IIncomeDailyDo
	WithContext(ctx context.Context) IIncomeDailyDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IIncomeDailyDo
	WriteDB() IIncomeDailyDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IIncomeDailyDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IIncomeDailyDo
	Not(conds ...gen.Condition) IIncomeDailyDo
	Or(conds ...gen.Condition) IIncomeDailyDo
	Select(conds ...field.Expr) IIncomeDailyDo
	Where(conds ...gen.Condition) IIncomeDailyDo
	Order(conds ...field.Expr) IIncomeDailyDo
	Distinct(cols ...field.Expr) IIncomeDailyDo
	Omit(cols ...field.Expr) IIncomeDailyDo
	Join(table schema.Tabler, on ...field.Expr) IIncomeDailyDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IIncomeDailyDo
	RightJoin(table schema.Tabler, on ...field.Expr) IIncomeDailyDo
	Group(cols ...field.Expr) IIncomeDailyDo
	Having(conds ...gen.Condition) IIncomeDailyDo
	Limit(limit int) IIncomeDailyDo
	Offset(offset int) IIncomeDailyDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IIncomeDailyDo
	Unscoped() IIncomeDailyDo
	Create(values ...*model.IncomeDaily) error
	CreateInBatches(values []*model.IncomeDaily, batchSize int) error
	Save(values ...*model.IncomeDaily) error
	First() (*model.IncomeDaily, error)
	Take() (*model.IncomeDaily, error)
	Last() (*model.IncomeDaily, error)
	Find() ([]*model.IncomeDaily, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.IncomeDaily, err error)
	FindInBatches(result *[]*model.IncomeDaily, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.IncomeDaily) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IIncomeDailyDo
	Assign(attrs ...field.AssignExpr) IIncomeDailyDo
	Joins(fields ...field.RelationField) IIncomeDailyDo
	Preload(fields ...field.RelationField) IIncomeDailyDo
	FirstOrInit() (*model.IncomeDaily, error)
	FirstOrCreate() (*model.IncomeDaily, error)
	FindByPage(offset int, limit int) (result []*model.IncomeDaily, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IIncomeDailyDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (i incomeDailyDo) Debug() IIncomeDailyDo {
	return i.withDO(i.DO.Debug())
}

func (i incomeDailyDo) WithContext(ctx context.Context) IIncomeDailyDo {
	return i.withDO(i.DO.WithContext(ctx))
}

func (i incomeDailyDo) ReadDB() IIncomeDailyDo {
	return i.Clauses(dbresolver.Read)
}

func (i incomeDailyDo) WriteDB() IIncomeDailyDo {
	return i.Clauses(dbresolver.Write)
}

func (i incomeDailyDo) Session(config *gorm.Session) IIncomeDailyDo {
	return i.withDO(i.DO.Session(config))
}

func (i incomeDailyDo) Clauses(conds ...clause.Expression) IIncomeDailyDo {
	return i.withDO(i.DO.Clauses(conds...))
}

func (i incomeDailyDo) Returning(value interface{}, columns ...string) IIncomeDailyDo {
	return i.withDO(i.DO.Returning(value, columns...))
}

func (i incomeDailyDo) Not(conds ...gen.Condition) IIncomeDailyDo {
	return i.withDO(i.DO.Not(conds...))
}

func (i incomeDailyDo) Or(conds ...gen.Condition) IIncomeDailyDo {
	return i.withDO(i.DO.Or(conds...))
}

func (i incomeDailyDo) Select(conds ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Select(conds...))
}

func (i incomeDailyDo) Where(conds ...gen.Condition) IIncomeDailyDo {
	return i.withDO(i.DO.Where(conds...))
}

func (i incomeDailyDo) Exists(subquery interface{ UnderlyingDB() *gorm.DB }) IIncomeDailyDo {
	return i.Where(field.CompareSubQuery(field.ExistsOp, nil, subquery.UnderlyingDB()))
}

func (i incomeDailyDo) Order(conds ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Order(conds...))
}

func (i incomeDailyDo) Distinct(cols ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Distinct(cols...))
}

func (i incomeDailyDo) Omit(cols ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Omit(cols...))
}

func (i incomeDailyDo) Join(table schema.Tabler, on ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Join(table, on...))
}

func (i incomeDailyDo) LeftJoin(table schema.Tabler, on ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.LeftJoin(table, on...))
}

func (i incomeDailyDo) RightJoin(table schema.Tabler, on ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.RightJoin(table, on...))
}

func (i incomeDailyDo) Group(cols ...field.Expr) IIncomeDailyDo {
	return i.withDO(i.DO.Group(cols...))
}

func (i incomeDailyDo) Having(conds ...gen.Condition) IIncomeDailyDo {
	return i.withDO(i.DO.Having(conds...))
}

func (i incomeDailyDo) Limit(limit int) IIncomeDailyDo {
	return i.withDO(i.DO.Limit(limit))
}

func (i incomeDailyDo) Offset(offset int) IIncomeDailyDo {
	return i.withDO(i.DO.Offset(offset))
}

func (i incomeDailyDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IIncomeDailyDo {
	return i.withDO(i.DO.Scopes(funcs...))
}

func (i incomeDailyDo) Unscoped() IIncomeDailyDo {
	return i.withDO(i.DO.Unscoped())
}

func (i incomeDailyDo) Create(values ...*model.IncomeDaily) error {
	if len(values) == 0 {
		return nil
	}
	return i.DO.Create(values)
}

func (i incomeDailyDo) CreateInBatches(values []*model.IncomeDaily, batchSize int) error {
	return i.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (i incomeDailyDo) Save(values ...*model.IncomeDaily) error {
	if len(values) == 0 {
		return nil
	}
	return i.DO.Save(values)
}

func (i incomeDailyDo) First() (*model.IncomeDaily, error) {
	if result, err := i.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.IncomeDaily), nil
	}
}

func (i incomeDailyDo) Take() (*model.IncomeDaily, error) {
	if result, err := i.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.IncomeDaily), nil
	}
}

func (i incomeDailyDo) Last() (*model.IncomeDaily, error) {
	if result, err := i.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.IncomeDaily), nil
	}
}

func (i incomeDailyDo) Find() ([]*model.IncomeDaily, error) {
	result, err := i.DO.Find()
	return result.([]*model.IncomeDaily), err
}

func (i incomeDailyDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.IncomeDaily, err error) {
	buf := make([]*model.IncomeDaily, 0, batchSize)
	err = i.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (i incomeDailyDo) FindInBatches(result *[]*model.IncomeDaily, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return i.DO.FindInBatches(result, batchSize, fc)
}

func (i incomeDailyDo) Attrs(attrs ...field.AssignExpr) IIncomeDailyDo {
	return i.withDO(i.DO.Attrs(attrs...))
}

func (i incomeDailyDo) Assign(attrs ...field.AssignExpr) IIncomeDailyDo {
	return i.withDO(i.DO.Assign(attrs...))
}

func (i incomeDailyDo) Joins(fields ...field.RelationField) IIncomeDailyDo {
	for _, _f := range fields {
		i = *i.withDO(i.DO.Joins(_f))
	}
	return &i
}

func (i incomeDailyDo) Preload(fields ...field.RelationField) IIncomeDailyDo {
	for _, _f := range fields {
		i = *i.withDO(i.DO.Preload(_f))
	}
	return &i
}

func (i incomeDailyDo) FirstOrInit() (*model.IncomeDaily, error) {
	if result, err := i.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.IncomeDaily), nil
	}
}

func (i incomeDailyDo) FirstOrCreate() (*model.IncomeDaily, error) {
	if result, err := i.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.IncomeDaily), nil
	}
}

func (i incomeDailyDo) FindByPage(offset int, limit int) (result []*model.IncomeDaily, count int64, err error) {
	result, err = i.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = i.Offset(-1).Limit(-1).Count()
	return
}

func (i incomeDailyDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = i.Count()
	if err != nil {
		return
	}

	err = i.Offset(offset).Limit(limit).Scan(result)
	return
}

func (i incomeDailyDo) Scan(result interface{}) (err error) {
	return i.DO.Scan(result)
}

func (i incomeDailyDo) Delete(models ...*model.IncomeDaily) (result gen.ResultInfo, err error) {
	return i.DO.Delete(models)
}

func (i *incomeDailyDo) withDO(do gen.Dao) *incomeDailyDo {
	i.DO = *do.(*gen.DO)
	return i
}
