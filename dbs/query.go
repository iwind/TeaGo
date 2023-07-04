package dbs

import (
	"database/sql"
	"errors"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

const (
	QueryActionFind           = 1
	QueryActionDelete         = 2
	QueryActionInsert         = 3
	QueryActionReplace        = 4
	QueryActionInsertOrUpdate = 5
	QueryActionUpdate         = 6
	QueryActionExec           = 7
)

const (
	QuerySubActionCount = 1
	QuerySubActionSum   = 2
	QuerySubActionMax   = 3
	QuerySubActionMin   = 4
	QuerySubActionAvg   = 5
)

const (
	QueryOrderDefault = 0
	QueryOrderAsc     = 1
	QueryOrderDesc    = 2
)

const (
	QueryJoinDefault = 0
	QueryJoinLeft    = 1
	QueryJoinRight   = 2
)

const (
	QueryLockShareMode = "LOCK IN SHARE MODE"
	QueryLockForUpdate = "FOR UPDATE"
)

const (
	QuerySqlCacheDefault = 0
	QuerySqlCacheOn      = 1
	QuerySqlCacheOff     = 2
)

const (
	QueryForAll     = ""
	QueryForJoin    = "JOIN"
	QueryForOrderBy = "ORDER BY"
	QueryForGroupBy = "GROUP BY"
)

// ErrNotFound 错误信息
var ErrNotFound = errors.New("record not found")

// SQL缓存
var sqlCacheMap = map[string]map[string]any{} // sql => { params:[]string{} sql:string }
var sqlCacheLocker = sync.RWMutex{}

type Query struct {
	db  *DB
	tx  *Tx
	dao *DAOObject

	model  *Model
	table  string
	pkName string

	canReuse bool // 是否可以重用的，如果是可以重用的使用PrepareOnce()来Prepare SQL，默认为true

	action    int
	subAction int

	sql     string
	noPk    bool
	attrs   *maps.OrderedMap
	wheres  []string
	havings []string
	orders  []QueryOrder
	groups  []QueryGroup

	limit  int64
	offset int64

	results       []string
	exceptResults []string
	joins         []QueryJoin
	partitions    []string
	useIndexes    []*QueryUseIndex

	sqlCache int
	lock     string

	savingFields    *maps.OrderedMap // 要插入或修改的字段列表
	replacingFields *maps.OrderedMap // 执行replace的字段列表

	debug bool

	filterFn func(one maps.Map) bool
	mapFn    func(one maps.Map) maps.Map
	slicePtr any

	namedParamPrefix string // 命名参数名前缀
	namedParams      map[string]any
	namedParamIndex  int

	params []any

	isSub bool // 是否为子查询
}

type QueryOrder struct {
	Field any
	Type  int
}

type QueryGroup struct {
	Field string
	Order int
}

type QueryJoin struct {
	DAO  *DAOObject
	Type int
	On   string
}

type QueryUseIndex struct {
	Keyword string
	For     string
	Indexes []string
}

func NewQuery(model any) *Query {
	var query = &Query{}
	query.init(model)
	return query
}

func (this *Query) init(model any) *Query {
	if model != nil {
		this.model = NewModel(model)
	}

	this.canReuse = true

	this.sqlCache = QuerySqlCacheDefault

	this.action = QueryActionFind
	this.pkName = "id"
	this.noPk = true

	this.limit = -1
	this.offset = -1
	this.debug = false

	this.attrs = maps.NewOrderedMap()
	this.savingFields = maps.NewOrderedMap()
	this.replacingFields = maps.NewOrderedMap()

	this.namedParams = map[string]any{}
	this.namedParamIndex = 0

	this.namedParamPrefix = ""

	return this
}

// DB 设置数据库实例
func (this *Query) DB(db *DB) *Query {
	this.db = db
	return this
}

// Tx 设置事务
func (this *Query) Tx(tx *Tx) *Query {
	this.tx = tx
	return this
}

// DAO 设置DAO
func (this *Query) DAO(dao *DAOObject) *Query {
	this.dao = dao
	return this
}

// Table 设置表名
func (this *Query) Table(table string) *Query {
	this.table = table
	return this
}

// Reuse 是否可以重用，对于根据参数变化而变化的查询，需要设置为false
func (this *Query) Reuse(canReuse bool) *Query {
	this.canReuse = canReuse
	return this
}

// State 设置状态查询快捷函数
// 相当于：Attr("state", state)
func (this *Query) State(state any) *Query {
	return this.Attr("state", state)
}

// NoPk 设置是否返回主键
// 默认find或findAll查询中返回主键，以便后续可以用此主键值操作对象
func (this *Query) NoPk(noPk bool) *Query {
	this.noPk = noPk
	return this
}

// PkName 设置主键名
// @TODO 支持联合主键
func (this *Query) PkName(pkName string) *Query {
	this.pkName = pkName
	return this
}

// Attr 设置查询的字段
func (this *Query) Attr(name string, value any) *Query {
	placeholder, isSlice := this.wrapAttr(value)
	if isSlice {
		this.attrs.Put(name, this.wrapKeyword(name)+" "+placeholder)
	} else {
		this.attrs.Put(name, this.wrapKeyword(name)+"="+placeholder)
	}

	return this
}

// Offset 设置偏移量
func (this *Query) Offset(offset int64) *Query {
	this.offset = offset
	return this
}

// Size 设置Limit条件，同limit()
func (this *Query) Size(size int64) *Query {
	return this.Limit(size)
}

// Limit 设置Limit条件，同size()
func (this *Query) Limit(size int64) *Query {
	this.limit = size
	return this
}

// Result 设置查询要返回的字段
// 字段名中支持星号(*)通配符
func (this *Query) Result(fields ...any) *Query {
	for _, field := range fields {
		if _, ok := field.(string); ok {
			this.results = append(this.results, field.(string))
		} else if _, ok := field.(*DBFunc); ok {
			this.results = append(this.results, field.(*DBFunc).prepareForQuery(this))
		} else if _, ok := field.(*Query); ok {
			sqlString, err := field.(*Query).AsSQL()
			if err != nil {
				logs.Errorf("%s", err.Error())
				return this
			}
			this.results = append(this.results, "("+sqlString+")")
		}
	}
	return this
}

// ResultExcept 设置查询不需要返回的字段
func (this *Query) ResultExcept(fields ...string) *Query {
	this.exceptResults = append(this.exceptResults, fields...)
	return this
}

// ResultPk 设置返回主键字段值
func (this *Query) ResultPk() *Query {
	return this.Result(this.pkName)
}

// Debug 设置是否开启调试模式
// 如果开启调试模式，会打印SQL语句
func (this *Query) Debug(debug bool) *Query {
	this.debug = debug
	return this
}

// Order 添加排序条件
func (this *Query) Order(field any, orderType int) *Query {
	this.orders = append(this.orders, QueryOrder{
		Field: field,
		Type:  orderType,
	})
	return this
}

// Asc 添加正序排序
func (this *Query) Asc(fields ...string) *Query {
	for _, field := range fields {
		this.Order(field, QueryOrderAsc)
	}
	return this
}

// AscPk 按照主键倒排序
func (this *Query) AscPk() *Query {
	this.Order(this.pkName, QueryOrderAsc)
	return this
}

// Desc 添加倒序排序
func (this *Query) Desc(fields ...string) *Query {
	for _, field := range fields {
		this.Order(field, QueryOrderDesc)
	}
	return this
}

// DescPk 按照主键倒排序
func (this *Query) DescPk() *Query {
	this.Order(this.pkName, QueryOrderDesc)
	return this
}

// Join 设置单个联合查询条件
func (this *Query) Join(dao any, joinType int, on string) *Query {
	this.joins = append(this.joins, QueryJoin{
		DAO:  dao.(DAOWrapper).Object(),
		Type: joinType,
		On:   on,
	})
	return this
}

// HasJoins 判断是否有联合查询条件
func (this *Query) HasJoins() bool {
	return len(this.joins) > 0
}

// Having 设置Having条件
func (this *Query) Having(cond string) *Query {
	this.havings = append(this.havings, cond)
	return this
}

// Where 设置where条件
// @TODO 支持Query、SQL
func (this *Query) Where(wheres ...string) *Query {
	this.wheres = append(this.wheres, wheres...)

	return this
}

// Gt 设置大于条件
func (this *Query) Gt(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + ">" + ":" + param)

	return this
}

// Gte 设置大于等于条件
func (this *Query) Gte(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + ">=" + ":" + param)

	return this
}

// Lt 设置小于条件
func (this *Query) Lt(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "<" + ":" + param)

	return this
}

// Lte 设置小于等于条件
func (this *Query) Lte(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "<=" + ":" + param)

	return this
}

// Neq 设置不等于条件
func (this *Query) Neq(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "!=" + ":" + param)

	return this
}

// Group 设置Group查询条件
func (this *Query) Group(field string, order ...int) *Query {
	realOrder := QueryOrderDefault
	if len(order) > 0 {
		realOrder = order[0]
	}

	var group = QueryGroup{
		Field: field,
		Order: realOrder,
	}
	this.groups = append(this.groups, group)
	return this
}

// Like 设置like查询条件
// 对表达式自动加上百分号， % ... %
func (this *Query) Like(field string, expr string) *Query {
	wrappedValue, _ := this.wrapAttr("%" + expr + "%")
	this.Where(this.wrapKeyword(field) + " LIKE " + wrappedValue)
	return this
}

// JSONContains JSON包含
func (this *Query) JSONContains(attr string, value any) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where("JSON_CONTAINS(" + this.wrapKeyword(attr) + ", :" + param + ")")

	return this
}

// SQLCache 是否开启SQL Cache
// 只有在SELECT时才有作用
func (this *Query) SQLCache(sqlCache int) *Query {
	this.sqlCache = sqlCache
	return this
}

// Lock 行锁定
func (this *Query) Lock(lock string) *Query {
	this.lock = lock
	return this
}

// UseIndex 使用索引
func (this *Query) UseIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "USE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// IgnoreIndex 屏蔽索引
func (this *Query) IgnoreIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "IGNORE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// ForceIndex 强制使用索引
func (this *Query) ForceIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "FORCE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// For 针对的操作
// 和 UserIndex, IgnoreIndex, ForceIndex 配合使用
func (this *Query) For(clause string) *Query {
	if len(this.useIndexes) == 0 {
		return this
	}

	lastIndex := this.useIndexes[len(this.useIndexes)-1]
	lastIndex.For = clause
	return this
}

// Partitions 指定分区
func (this *Query) Partitions(partitions ...string) *Query {
	this.partitions = append(this.partitions, partitions...)
	return this
}

// 分区片段SQL
func (this *Query) partitionsSQL() string {
	if len(this.partitions) > 0 {
		return " PARTITION(" + strings.Join(this.partitions, ", ") + ") "
	}
	return ""
}

// Pk 设置要查询的主键值
func (this *Query) Pk(pks ...any) *Query {
	var realPks = []any{}
	for _, pk := range pks {
		var value = reflect.ValueOf(pk)
		if value.Kind() == reflect.Slice {
			count := value.Len()
			for i := 0; i < count; i++ {
				realPks = append(realPks, value.Index(i).Interface())
			}
		} else {
			realPks = append(realPks, pk)
		}
	}

	this.Attr(this.pkName, realPks)

	return this
}

// Between 设置between条件
func (this *Query) Between(field string, min any, max any) *Query {
	minValue, _ := this.wrapAttr(min)
	maxValue, _ := this.wrapAttr(max)
	this.Where(field + " BETWEEN " + minValue + " AND " + maxValue)
	return this
}

// Attrs 设置一组查询的字段
func (this *Query) Attrs(attrs maps.Map) *Query {
	for key, value := range attrs {
		this.Attr(types.String(key), value)
	}
	return this
}

// Increase 增加某个字段的数值
func (this *Query) Increase(field string, count int) *Query {
	this.savingFields.Put(field, this.wrapKeyword(field)+"+"+this.wrapValue(count))
	return this
}

// Decrease 减少某个字段的数值
func (this *Query) Decrease(field string, count int) *Query {
	this.savingFields.Put(field, this.wrapKeyword(field)+"-"+this.wrapValue(count))
	return this
}

// Set 设置字段值，以便用于删除和修改操作
func (this *Query) Set(field string, value any) *Query {
	this.savingFields.Put(field, this.wrapValue(value))
	return this
}

// Sets 设置一组字段值，以便用于删除和修改操作
func (this *Query) Sets(values map[string]any) *Query {
	for field, value := range values {
		this.savingFields.Put(field, this.wrapValue(value))
	}
	return this
}

// Param 设定查询语句中的参数值
// 只有指定where和sql后，才能使用该方法
func (this *Query) Param(name string, value any) *Query {
	this.namedParams[name] = value
	return this
}

// SQL 指定SQL语句
func (this *Query) SQL(sql string) *Query {
	this.sql = sql
	return this
}

// Filter 指定筛选程序
func (this *Query) Filter(filterFn func(one maps.Map) bool) *Query {
	this.filterFn = filterFn
	return this
}

// Map 指定映射程序
func (this *Query) Map(mapFn func(one maps.Map) maps.Map) *Query {
	this.mapFn = mapFn
	return this
}

// Slice 设置返回的Slice指针
// 在FindAll()方法中生效
func (this *Query) Slice(slicePtr any) *Query {
	this.slicePtr = slicePtr
	return this
}

// AsSQL 将查询转换为SQL语句
func (this *Query) AsSQL() (string, error) {
	// SQL
	var sqlString = this.sql

	if len(sqlString) == 0 {
		if len(this.table) == 0 {
			return "", errors.New("you need specify a table name")
		}

		if this.action == QueryActionFind {
			var resultString = "*"

			if len(this.results) > 0 {
				if !this.noPk && len(this.pkName) > 0 && !stringutil.Contains(this.results, this.pkName) {
					this.results = append(this.results, this.pkName)
				}
				var newResults = []string{}
				for _, result := range this.results {
					// check except fields ...
					if len(this.exceptResults) > 0 && lists.ContainsString(this.exceptResults, result) {
						continue
					}

					newResults = append(newResults, this.wrapKeyword(result))
				}
				if len(newResults) > 0 {
					resultString = strings.Join(newResults, ", ")
				}
			} else if len(this.exceptResults) > 0 && this.dao != nil && len(this.dao.fields) > 0 {
				var newResults = []string{}
				for _, fieldObj := range this.dao.fields {
					if lists.ContainsString(this.exceptResults, fieldObj.Name) {
						continue
					}
					newResults = append(newResults, this.wrapKeyword(fieldObj.Name))
				}
				if len(newResults) > 0 { // make sure we have result fields
					resultString = strings.Join(newResults, ", ")
				}
			}

			sqlString = "SELECT"
			if this.sqlCache == QuerySqlCacheOn {
				sqlString += " SQL_CACHE "
			} else if this.sqlCache == QuerySqlCacheOff {
				sqlString += " SQL_NO_CACHE "
			}

			sqlString += "\n  " + resultString + "\n FROM "
			sqlString += this.wrapTable(this.table)

			// use indexes
			if len(this.useIndexes) > 0 {
				for _, useIndex := range this.useIndexes {
					sqlString += "\n  " + useIndex.Keyword + " INDEX"
					if len(useIndex.For) > 0 {
						sqlString += " FOR " + useIndex.For
					}
					quotedIndexes := []string{}
					for _, indexName := range useIndex.Indexes {
						quotedIndexes = append(quotedIndexes, this.wrapKeyword(indexName))
					}
					sqlString += " (" + strings.Join(quotedIndexes, ", ") + ")"
				}
			}

			// joins
			if len(this.joins) > 0 {
				for _, join := range this.joins {
					if join.Type == QueryJoinDefault {
						sqlString += ", " + this.wrapTable(join.DAO.Table)
						if len(join.On) > 0 {
							this.wheres = append(this.wheres, join.On)
						}
					}
				}

				for _, join := range this.joins {
					if join.Type == QueryJoinDefault {
						continue
					}
					switch join.Type {
					case QueryJoinLeft:
						sqlString += "\n LEFT JOIN " + this.wrapTable(join.DAO.Table)
					case QueryJoinRight:
						sqlString += "\n RIGHT JOIN " + this.wrapTable(join.DAO.Table)
					}
					sqlString += this.partitionsSQL()
					if len(join.On) > 0 {
						sqlString += " ON " + join.On
					}
				}
			} else {
				sqlString += this.partitionsSQL()
			}
		} else if this.action == QueryActionDelete {
			sqlString = "DELETE FROM " + this.wrapTable(this.table) + "\n " + this.partitionsSQL()
		} else if this.action == QueryActionUpdate {
			sqlString = "UPDATE " + this.wrapTable(this.table) + "\n " + this.partitionsSQL() + "SET"
			if this.savingFields.Len() > 0 {
				var mapping = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field any, value any) {
					mapping = append(mapping, this.wrapKeyword(field.(string))+"="+value.(string))
				})
				sqlString += " " + strings.Join(mapping, ", ")
			}

		} else if this.action == QueryActionInsert {
			sqlString = "INSERT INTO " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field any, value any) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sqlString += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}
		} else if this.action == QueryActionReplace {
			sqlString = "REPLACE " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field any, value any) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sqlString += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}
		} else if this.action == QueryActionInsertOrUpdate {
			sqlString = "INSERT INTO " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field any, value any) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sqlString += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}

			sqlString += "\nON DUPLICATE KEY UPDATE\n"

			if this.replacingFields.Len() > 0 {
				var mapping = []string{}
				this.replacingFields.SortKeys()
				this.replacingFields.Range(func(field any, value any) {
					mapping = append(mapping, this.wrapKeyword(field.(string))+"="+value.(string))
				})
				sqlString += strings.Join(mapping, ", ")
			}
		}
	} else {
		sqlString = this.sql
	}

	// attrs
	var wheres = []string{}
	if this.attrs.Len() > 0 {
		this.attrs.Range(func(_ any, placeholder any) {
			wheres = append(wheres, placeholder.(string))
		})
	}

	// where
	if len(this.wheres) > 0 {
		wheres = append(wheres, this.wheres...)
	}
	if this.action != QueryActionInsert && this.action != QueryActionReplace && this.action != QueryActionInsertOrUpdate && len(wheres) > 0 {
		sqlString += "\n WHERE " + strings.Join(wheres, " AND ")
	}

	// group
	if this.action == QueryActionFind && len(this.groups) > 0 {
		var groupStrings = []string{}
		for _, group := range this.groups {
			groupStrings = append(groupStrings, this.wrapKeyword(group.Field)+" "+this.orderCode(group.Order))
		}
		sqlString += "\n GROUP BY " + strings.Join(groupStrings, ", ")
	}

	// having
	if len(this.havings) > 0 {
		sqlString += "\n HAVING " + strings.Join(this.havings, " AND ")
	}

	// orders
	if len(this.orders) > 0 {
		var orderStrings = []string{}
		for _, order := range this.orders {
			var fieldString = ""
			if _, ok := order.Field.(string); ok {
				fieldString = order.Field.(string)
			} else if _, ok := order.Field.(SQL); ok {
				fieldString = string(order.Field.(SQL))
			} else if _, ok := order.Field.(*Query); ok {
				fieldString, _ = order.Field.(*Query).AsSQL()
			} else if _, ok := order.Field.(*DBFunc); ok {
				fieldString = order.Field.(*DBFunc).prepareForQuery(this)
			} else {
				return "", errors.New("invalid order field")
			}
			orderStrings = append(orderStrings, this.wrapKeyword(fieldString)+" "+this.orderCode(order.Type))
		}
		sqlString += "\n ORDER BY " + strings.Join(orderStrings, ", ")
	}

	// limit & offset
	if this.subAction == 0 {
		if this.limit > -1 {
			if this.offset > -1 {
				offsetValue, _ := this.wrapAttr(this.offset)
				limitValue, _ := this.wrapAttr(this.limit)
				sqlString += "\n LIMIT " + types.String(offsetValue) + ", " + types.String(limitValue)
			} else {
				limitValue, _ := this.wrapAttr(this.limit)
				sqlString += "\n LIMIT " + types.String(limitValue)
			}
		}
	}

	// JOIN
	if len(this.joins) > 0 {
		reg, _ := stringutil.RegexpCompile("\\b" + "self" + "\\s*\\.")
		sqlString = reg.ReplaceAllString(sqlString, this.wrapTable(this.table)+".")

		reg, _ = stringutil.RegexpCompile("\\b" + this.model.Type.Name() + "\\s*\\.")
		sqlString = reg.ReplaceAllString(sqlString, this.wrapTable(this.table)+".")

		for _, join := range this.joins {
			modelName := join.DAO.modelWrapper.Type.Name()
			reg, _ := stringutil.RegexpCompile("\\b" + modelName + "\\s*\\.")
			sqlString = reg.ReplaceAllString(sqlString, this.wrapTable(join.DAO.Table)+".")
		}
	}

	// lock
	if this.action == QueryActionFind && len(this.lock) > 0 {
		sqlString += "\n " + this.lock
	}

	// 处理:NamedParam
	var resultSQL = sqlString
	if !this.isSub {
		this.params = []any{}
		resultSQL = this.parsePlaceholders(sqlString)
	}

	// debug
	if this.debug {
		logs.Debugf("SQL:" + sqlString)
		logs.Debugf("params:%#v", this.namedParams)
	}

	return resultSQL, nil
}

// FindOnes 查找一组数据，返回map数据
func (this *Query) FindOnes() (ones []maps.Map, columnNames []string, err error) {
	this.action = QueryActionFind
	sqlString, err := this.AsSQL()
	if err != nil {
		return nil, nil, err
	}

	if this.canReuse {
		stmt, cached, prepareErr := this.executor().PrepareOnce(sqlString)
		if prepareErr != nil {
			return nil, nil, prepareErr
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		ones, columnNames, err = stmt.FindOnes(this.params...)
	} else {
		ones, columnNames, err = this.executor().FindOnes(sqlString, this.params...)
	}
	if err != nil {
		return nil, nil, err
	}

	// 执行 filterFn 和 mapFn
	if this.filterFn != nil || this.mapFn != nil {
		var results = []maps.Map{}
		for _, one := range ones {
			if this.filterFn != nil {
				if !this.filterFn(one) {
					continue
				}
			}
			if this.mapFn != nil {
				one = this.mapFn(one)
			}
			results = append(results, one)
		}
		ones = results
	}

	return
}

// FindOne 查找一行数据
func (this *Query) FindOne() (results maps.Map, columnNames []string, err error) {
	this.limit = 1
	if this.offset < 0 {
		this.offset = 0
	}
	ones, columnNames, err := this.FindOnes()
	if err != nil {
		return nil, nil, err
	}

	if len(ones) == 0 {
		return nil, nil, nil
	}

	return ones[0], columnNames, nil
}

// FindAll 查询一组数据， 并返回模型数据
func (this *Query) FindAll() ([]any, error) {
	var ones, _, err = this.FindOnes()
	if err != nil {
		return nil, err
	}

	var results = []any{}
	if this.slicePtr == nil {
		for _, one := range ones {
			var value = this.copyModelValue(this.model.Type, one)
			results = append(results, value)
		}
	} else { // 将模型对象存入到指定的Slice中
		ptrValue := reflect.ValueOf(this.slicePtr)
		sliceValue := reflect.Indirect(ptrValue)
		for _, one := range ones {
			var value = this.copyModelValue(this.model.Type, one)
			sliceValue = reflect.Append(sliceValue, reflect.ValueOf(value))
		}
		ptrValue.Elem().Set(sliceValue)
	}

	return results, nil
}

// Find 查询单条数据，返回模型对象
func (this *Query) Find() (any, error) {
	this.Limit(1)
	var results, err = this.FindAll()
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	return results[0], nil
}

// FindCol 查询单个字段值
func (this *Query) FindCol(defaultValue any) (any, error) {
	this.noPk = true
	var one, columnNames, err = this.FindOne()
	if err != nil {
		return defaultValue, err
	}
	if len(one) == 0 {
		return defaultValue, nil
	}
	var value, ok = one[columnNames[0]]
	if !ok {
		return defaultValue, nil
	}
	if value == nil {
		return defaultValue, nil
	}
	return value, nil
}

// FindStringCol 查询单个字段值并返回字符串
func (this *Query) FindStringCol(defaultValue string) (string, error) {
	col, err := this.FindCol(defaultValue)
	return types.String(col), err
}

// FindBoolCol 查询单个字段值并返回bool值
func (this *Query) FindBoolCol() (bool, error) {
	col, err := this.FindCol("0")
	return types.String(col) == "1", err
}

// FindBytesCol 查询单个字段值并返回字节Slice
func (this *Query) FindBytesCol() ([]byte, error) {
	col, err := this.FindCol("")
	return []byte(types.String(col)), err
}

// FindJSONCol 查询单个字段值并返回JSON slice
func (this *Query) FindJSONCol() ([]byte, error) {
	col, err := this.FindCol("")
	return JSON(types.String(col)), err
}

// FindIntCol 查询某个字段值并返回整型
func (this *Query) FindIntCol(defaultValue int) (int, error) {
	col, err := this.FindCol(defaultValue)
	return types.Int(col), err
}

// FindInt64Col 查询某个字段值并返回64位整型
func (this *Query) FindInt64Col(defaultValue int64) (int64, error) {
	col, err := this.FindCol(defaultValue)
	return types.Int64(col), err
}

// FindFloat64Col 查询某个字段值并返回64位浮点型
func (this *Query) FindFloat64Col(defaultValue float64) (float64, error) {
	col, err := this.FindCol(defaultValue)
	return types.Float64(col), err
}

// FindFloat32Col 查询某个字段值并返回32位浮点型
func (this *Query) FindFloat32Col(defaultValue float32) (float32, error) {
	col, err := this.FindCol(defaultValue)
	return types.Float32(col), err
}

// Exist 判断记录是否存在
func (this *Query) Exist() (bool, error) {
	var one, _, err = this.ResultPk().FindOne()
	if err != nil {
		return false, err
	}
	return len(one) > 0, nil
}

// Count 执行COUNT查询
func (this *Query) Count() (int64, error) {
	return this.CountAttr("*")
}

// CountInt 执行Count查询并返回int
func (this *Query) CountInt() (int, error) {
	count, err := this.Count()
	return int(count), err
}

// CountAttr 对某个字段进行COUNT查询
func (this *Query) CountAttr(attr string) (int64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionCount
	this.NoPk(true)

	this.results = []string{"COUNT(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(0)
	if err != nil {
		return 0, err
	}

	return types.Int64(value), err
}

// Sum 执行SUM查询
func (this *Query) Sum(attr string, defaultValue float64) (float64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionSum
	this.NoPk(true)

	this.results = []string{"SUM(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Float64(value), err
}

// SumInt 执行SUM查询，并返回Int
func (this *Query) SumInt(attr string, defaultValue int) (int, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionSum
	this.NoPk(true)

	this.results = []string{"SUM(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Int(value), err
}

// SumInt64 执行SUM查询，并返回Int64
func (this *Query) SumInt64(attr string, defaultValue int64) (int64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionSum
	this.NoPk(true)

	this.results = []string{"SUM(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Int64(value), err
}

// Min 执行MIN查询
func (this *Query) Min(attr string, defaultValue float64) (float64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionMin
	this.NoPk(true)

	this.results = []string{"MIN(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Float64(value), err
}

// Max 执行MAX查询
func (this *Query) Max(attr string, defaultValue float64) (float64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionMax
	this.NoPk(true)

	this.results = []string{"MAX(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Float64(value), err
}

// MaxInt64 执行MAX查询
func (this *Query) MaxInt64(attr string, defaultValue int64) (int64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionMax
	this.NoPk(true)

	this.results = []string{"MAX(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Int64(value), err
}

// MaxInt 执行MAX查询
func (this *Query) MaxInt(attr string, defaultValue int) (int, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionMax
	this.NoPk(true)

	this.results = []string{"MAX(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Int(value), err
}

// Avg 执行AVG查询
func (this *Query) Avg(attr string, defaultValue float64) (float64, error) {
	this.action = QueryActionFind
	this.subAction = QuerySubActionAvg
	this.NoPk(true)

	this.results = []string{"AVG(" + this.wrapKeyword(attr) + ")"}
	var value, err = this.FindCol(defaultValue)
	if err != nil {
		return 0, err
	}

	if value == nil {
		return defaultValue, nil
	}
	return types.Float64(value), err
}

// Exec 执行查询
func (this *Query) Exec() (*Result, error) {
	this.action = QueryActionExec

	sqlString, err := this.AsSQL()
	if err != nil {
		return nil, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return nil, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return nil, err
	}

	return NewResult(result), nil
}

// Replace 执行REPLACE
func (this *Query) Replace() (rowsAffected int64, lastInsertId int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, 0, errors.New("[Query.Replace()]Replacing fields should be set")
	}

	this.action = QueryActionReplace
	sqlString, err := this.AsSQL()
	if err != nil {
		return 0, 0, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return 0, 0, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return 0, 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return rows, 0, err
	}
	return rows, lastId, err
}

// Insert 执行INSERT
func (this *Query) Insert() (lastInsertId int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, errors.New("[Query.Insert()]inserting fields should be set")
	}

	this.action = QueryActionInsert
	sqlString, err := this.AsSQL()
	if err != nil {
		return 0, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return 0, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return 0, err
	}

	lastInsertId, err = result.LastInsertId()
	if err != nil {
		return
	}

	// 事件通知
	err = this.dao.NotifyInsert()
	return lastInsertId, err
}

// Update 执行UPDATE
func (this *Query) Update() (rowsAffected int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, errors.New("[Query.Update()]updating fields should be set")
	}

	this.action = QueryActionUpdate
	sqlString, err := this.AsSQL()
	if err != nil {
		return 0, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return 0, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return 0, err
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return rowsAffected, err
	}

	// 事件通知
	err = this.dao.NotifyUpdate()
	return rowsAffected, err
}

// UpdateQuickly 执行UPDATE
func (this *Query) UpdateQuickly() error {
	if this.savingFields.Len() == 0 {
		return errors.New("[Query.Update()]updating fields should be set")
	}

	this.action = QueryActionUpdate
	sqlString, err := this.AsSQL()
	if err != nil {
		return err
	}

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		_, err = stmt.Exec(this.params...)
	} else {
		_, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return err
	}

	// 事件通知
	err = this.dao.NotifyUpdate()
	return err
}

// InsertOrUpdate 插入或更改
// 依据要插入的数据中的unique键来决定是插入数据还是替换数据
func (this *Query) InsertOrUpdate(insertingValues maps.Map, updatingValues maps.Map) (rowsAffected int64, lastInsertId int64, err error) {
	if len(insertingValues) == 0 {
		return 0, 0, errors.New("[Query.InsertOrUpdate()]inserting values should be set")
	}
	if len(updatingValues) == 0 {
		return 0, 0, errors.New("[Query.InsertOrUpdate()]updating values should be set")
	}

	// insert的值
	for field, value := range insertingValues {
		this.savingFields.Put(field, this.wrapValue(value))
	}

	// replace的值
	for field, value := range updatingValues {
		this.replacingFields.Put(field, this.wrapValue(value))
	}

	this.action = QueryActionInsertOrUpdate
	sqlString, err := this.AsSQL()
	if err != nil {
		return 0, 0, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return 0, 0, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return 0, 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	lastId, err := result.LastInsertId()
	if err != nil {
		return rows, 0, err
	}

	if lastId > 0 {
		err = this.dao.NotifyInsert()
	} else {
		err = this.dao.NotifyUpdate()
	}

	return rows, lastId, err
}

// InsertOrUpdateQuickly 插入或更改
// 依据要插入的数据中的unique键来决定是插入数据还是替换数据
func (this *Query) InsertOrUpdateQuickly(insertingValues maps.Map, updatingValues maps.Map) error {
	if len(insertingValues) == 0 {
		return errors.New("[Query.InsertOrUpdate()]inserting values should be set")
	}
	if updatingValues == nil || len(updatingValues) == 0 {
		return errors.New("[Query.InsertOrUpdate()]updating values should be set")
	}

	// insert的值
	for field, value := range insertingValues {
		this.savingFields.Put(field, this.wrapValue(value))
	}

	// replace的值
	for field, value := range updatingValues {
		this.replacingFields.Put(field, this.wrapValue(value))
	}

	this.action = QueryActionInsertOrUpdate
	sqlString, err := this.AsSQL()
	if err != nil {
		return err
	}

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		_, err = stmt.Exec(this.params...)
	} else {
		_, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return err
	}

	err = this.dao.NotifyUpdate()
	if err != nil {
		return err
	}

	return nil
}

// Delete 执行DELETE
func (this *Query) Delete() (rowsAffected int64, err error) {
	this.action = QueryActionDelete
	sqlString, err := this.AsSQL()
	if err != nil {
		return 0, err
	}

	var result sql.Result

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return 0, err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		result, err = stmt.Exec(this.params...)
	} else {
		result, err = this.executor().Exec(sqlString, this.params...)
	}
	if err != nil {
		return 0, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	// 事件通知
	err = this.dao.NotifyDelete()
	if err != nil {
		return rows, err
	}

	return rows, nil
}

// DeleteQuickly 删除Delete，但不返回影响的行数
func (this *Query) DeleteQuickly() error {
	this.action = QueryActionDelete
	sqlString, err := this.AsSQL()
	if err != nil {
		return err
	}

	if this.canReuse {
		var stmt *Stmt
		var cached bool
		stmt, cached, err = this.executor().PrepareOnce(sqlString)
		if err != nil {
			return err
		}
		if !cached {
			defer func() {
				_ = stmt.Close()
			}()
		}

		_, err = stmt.Exec(this.params...)
	} else {
		_, err = this.executor().Exec(sqlString, this.params...)
	}
	return err
}

func (this *Query) stringValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case bool:
		if v {
			return 1
		}
		return 0
	case []byte:
		return string(v)
	case JSON:
		return string(v)

		// TODO 基础数据直接返回
	}

	return types.String(value)
}

// 包装值
func (this *Query) wrapAttr(value any) (placeholder string, isArray bool) {
	switch value1 := value.(type) {
	case SQL:
		return string(value1), false
	case *DBFunc:
		return value1.prepareForQuery(this), false
	case *Query:
		value1.isSub = true
		value1.sqlCache = QuerySqlCacheDefault // 子查询不支持SQL_CACHE
		sqlString, err := value1.AsSQL()
		if err != nil {
			logs.Errorf("%s", err.Error())
			return
		}

		for paramName, paramValue := range value1.namedParams {
			this.namedParams[paramName] = paramValue
			this.namedParamIndex++
		}

		return "IN (" + sqlString + ")", true
	case *lists.List:
		return this.wrapAttr(value1.Slice)
	case JSON:
		var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
		this.namedParams[param] = string(value1)
		this.namedParamIndex++
		return ":" + param, false
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, string:
		var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
		this.namedParams[param] = value1
		this.namedParamIndex++
		return ":" + param, false
	}

	// slice
	var valueType = reflect.TypeOf(value)
	if valueType.Kind() == reflect.Slice {
		var params = []string{}
		var reflectValue = reflect.ValueOf(value)
		var countElements = reflectValue.Len()
		for i := 0; i < countElements; i++ {
			var v, _ = this.wrapAttr(reflectValue.Index(i).Interface())
			params = append(params, v)
		}
		countParams := len(params)
		if countParams == 1 {
			return params[0], false
		} else if countParams > 1 {
			return "IN (" + strings.Join(params, ", ") + ")", true
		}
		return "IN ()", true
	}

	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++
	return ":" + param, false
}

// 包装值
func (this *Query) wrapValue(value any) (placeholder string) {
	if value == nil {
		value = ""
	}

	switch v := value.(type) {
	case SQL:
		return string(v)
	case JSON:
		value = v.String()
	case []byte:
		value = string(v)
	case int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64, float32, float64, string:
	default:
		value = this.stringValue(value)
	}

	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++
	return ":" + param
}

// 关键词包装
func (this *Query) wrapKeyword(keyword string) string {
	if this.db == nil {
		logs.Errorf("[Query.wrapKeyword()]query.db should be not nil")
		return "\"" + keyword + "\""
	}
	if !this.isKeyword(keyword) {
		return keyword
	}
	switch this.db.Driver() {
	case "mysql":
		if len(this.joins) > 0 {
			return "`" + this.table + "`." + "`" + keyword + "`"
		}
		return "`" + keyword + "`"
	case "mssql":
		if len(this.joins) > 0 {
			return "[" + this.table + "]." + "[" + keyword + "]"
		}
		return "[" + keyword + "]"
	}
	if len(this.joins) > 0 {
		return "\"" + this.table + "\"." + "\"" + keyword + "\""
	}
	return "\"" + keyword + "\""
}

// 关键词包装
func (this *Query) wrapTable(keyword string) string {
	if this.db == nil {
		logs.Errorf("[Query.wrapKeyword()]query.db should be not nil")
		return "\"" + keyword + "\""
	}
	if !this.isKeyword(keyword) {
		return keyword
	}
	switch this.db.Driver() {
	case "mysql":
		return "`" + keyword + "`"
	case "mssql":
		return "[" + keyword + "]"
	}
	return "\"" + keyword + "\""
}

// 排序SQL代码
func (this *Query) orderCode(order int) string {
	if order == QueryOrderAsc {
		return "ASC"
	}
	if order == QueryOrderDesc {
		return "DESC"
	}
	return ""
}

// 拷贝模型值
func (this *Query) copyModelValue(valueType reflect.Type, data maps.Map) any {
	var pointerValue = reflect.New(valueType)
	var value = reflect.Indirect(pointerValue)
	for index, fieldName := range this.model.Fields {
		var fieldData, ok = data[fieldName]
		if !ok {
			continue
		}
		if fieldData == nil {
			continue
		}
		var fieldValue = value.Field(index)
		switch fieldValue.Kind() {
		case reflect.Slice:
			var convertedData = this.model.convertValue(fieldData, this.model.Kinds[index])
			if convertedData != nil {
				stringValue, isString := convertedData.(string)
				if isString {
					fieldValue.Set(reflect.ValueOf([]byte(stringValue)))
				}
				bytesValue, isBytes := convertedData.([]byte)
				if isBytes {
					fieldValue.Set(reflect.ValueOf(bytesValue))
				}
			}
		default:
			fieldValue.Set(reflect.ValueOf(this.model.convertValue(fieldData, this.model.Kinds[index])))
		}
	}
	return pointerValue.Interface()
}

// 获取Executor
func (this *Query) executor() SQLExecutor {
	if this.tx != nil {
		return this.tx
	}
	return this.db
}

// 判断某个字符串是否为关键词
func (this *Query) isKeyword(s string) bool {
	for _, r := range s {
		if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		return false
	}
	return true
}

// 分析语句中的占位
func (this *Query) parsePlaceholders(sqlString string) string {
	if len(sqlString) < 1024 {
		sqlCacheLocker.RLock()
		cache, ok := sqlCacheMap[sqlString]
		if ok {
			for _, param := range cache["params"].([]string) {
				value, ok2 := this.namedParams[param]
				if ok2 {
					this.params = append(this.params, value)
				} else {
					this.params = append(this.params, nil)
				}
			}

			sqlCacheLocker.RUnlock()
			return cache["sql"].(string)
		}
		sqlCacheLocker.RUnlock()
	}

	var word = []rune{}
	var isStarted = false
	var result = []rune{}
	var paramNames = []string{}
	for _, r := range sqlString {
		if isStarted {
			if r == '_' || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				word = append(word, r)
				continue
			} else {
				if len(word) > 0 {
					// word结束了
					paramNames = append(paramNames, string(word))
					value, ok := this.namedParams[string(word)]
					if ok {
						this.params = append(this.params, value)
					} else {
						this.params = append(this.params, nil)
					}
					result = append(result, '?')
				} else {
					result = append(result, ':') // 还原
				}

				isStarted = false
				word = nil
			}
		}
		if r == ':' {
			isStarted = true
		} else {
			result = append(result, r)
		}
	}

	// 最后一个word
	if isStarted {
		if len(word) > 0 {
			paramNames = append(paramNames, string(word))
			value, ok := this.namedParams[string(word)]
			if ok {
				this.params = append(this.params, value)
			} else {
				this.params = append(this.params, nil)
			}
			result = append(result, '?')
		} else {
			result = append(result, ':') // 还原
		}
	}

	if len(sqlString) < 1024 {
		sqlCacheLocker.Lock()

		// 防止过载
		if len(sqlCacheMap) > 100_000 {
			var l = len(sqlCacheMap) / 3
			for k := range sqlCacheMap {
				if l >= 0 {
					delete(sqlCacheMap, k)
				} else {
					break
				}
				l--
			}
		}

		sqlCacheMap[sqlString] = map[string]any{
			"sql":    string(result),
			"params": paramNames,
		}
		sqlCacheLocker.Unlock()
	}

	return string(result)
}
