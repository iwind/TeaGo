package dbs

import (
	"errors"
	"fmt"
	"github.com/iwind/TeaGo/lists"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"github.com/iwind/TeaGo/utils/string"
	"reflect"
	"strconv"
	"strings"
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

// 错误信息
var ErrNotFound = errors.New("record not found")

type Query struct {
	db *DB

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

	limit  int
	offset int

	results    []string
	joins      []QueryJoin
	partitions []string
	useIndexes []*QueryUseIndex

	sqlCache int
	lock     string

	savingFields    *maps.OrderedMap // 要插入或修改的字段列表
	replacingFields *maps.OrderedMap // 执行replace的字段列表

	debug bool

	filterFn func(one maps.Map) bool
	mapFn    func(one maps.Map) maps.Map
	slicePtr interface{}

	namedParamPrefix string // 命名参数名前缀
	namedParams      map[string]interface{}
	namedParamIndex  int

	params []interface{}

	isSub bool // 是否为子查询
}

type QueryOrder struct {
	Field interface{}
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

func NewQuery(model interface{}) *Query {
	var query = &Query{}
	query.init(model)
	return query
}

func (this *Query) init(model interface{}) *Query {
	if model != nil {
		this.model = NewModel(model)
	}

	this.canReuse = true

	this.action = QueryActionFind
	this.pkName = "id"
	this.noPk = true

	this.limit = -1
	this.offset = -1
	this.debug = false

	this.params = []interface{}{}
	this.attrs = maps.NewOrderedMap()
	this.wheres = []string{}
	this.havings = []string{}
	this.orders = []QueryOrder{}
	this.groups = []QueryGroup{}
	this.results = []string{}
	this.joins = []QueryJoin{}
	this.partitions = []string{}
	this.useIndexes = []*QueryUseIndex{}
	this.savingFields = maps.NewOrderedMap()
	this.replacingFields = maps.NewOrderedMap()

	this.namedParams = map[string]interface{}{}
	this.namedParamIndex = 0

	this.namedParamPrefix = ""

	return this
}

// 设置数据库实例
func (this *Query) DB(db *DB) *Query {
	this.db = db
	return this
}

// 设置表名
func (this *Query) Table(table string) *Query {
	this.table = table
	return this
}

// 是否可以重用，对于根据参数变化而变化的查询，需要设置为false
func (this *Query) Reuse(canReuse bool) *Query {
	this.canReuse = canReuse
	return this
}

// 设置状态查询
// 相当于：Attr("state", state)
func (this *Query) State(state interface{}) *Query {
	return this.Attr("state", state)
}

// 设置是否返回主键
// 默认find或findAll查询中返回主键，以便后续可以用此主键值操作对象
func (this *Query) NoPk(noPk bool) *Query {
	this.noPk = noPk
	return this
}

// 设置主键名
// @TODO 支持联合主键
func (this *Query) PkName(pkName string) *Query {
	this.pkName = pkName
	return this
}

// 设置查询的字段
func (this *Query) Attr(name string, value interface{}) *Query {
	placeholder, isSlice := this.wrapAttr(value)
	if isSlice {
		this.attrs.Put(name, this.wrapKeyword(name)+" "+placeholder)
	} else {
		this.attrs.Put(name, this.wrapKeyword(name)+"="+placeholder)
	}

	return this
}

// 设置偏移量
func (this *Query) Offset(offset int) *Query {
	this.offset = offset
	return this
}

// 设置Limit条件，同limit()
func (this *Query) Size(size int) *Query {
	return this.Limit(size)
}

// 设置Limit条件，同size()
func (this *Query) Limit(size int) *Query {
	this.limit = size
	return this
}

// 设置查询要返回的字段
// 字段名中支持星号(*)通配符
func (this *Query) Result(fields ...interface{}) *Query {
	for _, field := range fields {
		if _, ok := field.(string); ok {
			this.results = append(this.results, field.(string))
		} else if _, ok := field.(*DBFunc); ok {
			this.results = append(this.results, field.(*DBFunc).prepareForQuery(this))
		} else if _, ok := field.(*Query); ok {
			sql, err := field.(*Query).AsSQL()
			if err != nil {
				logs.Errorf("%s", err.Error())
				return this
			}
			this.results = append(this.results, "("+sql+")")
		}
	}
	return this
}

// 设置返回主键字段值
func (this *Query) ResultPk() *Query {
	return this.Result(this.pkName)
}

// 设置是否开启调试模式
// 如果开启调试模式，会打印SQL语句
func (this *Query) Debug(debug bool) *Query {
	this.debug = debug
	return this
}

// 添加排序条件
func (this *Query) Order(field interface{}, orderType int) *Query {
	this.orders = append(this.orders, QueryOrder{
		Field: field,
		Type:  orderType,
	})
	return this
}

// 添加正序排序
func (this *Query) Asc(fields ...string) *Query {
	for _, field := range fields {
		this.Order(field, QueryOrderAsc)
	}
	return this
}

// 按照主键倒排序
func (this *Query) AscPk() *Query {
	this.Order(this.pkName, QueryOrderAsc)
	return this
}

// 添加倒序排序
func (this *Query) Desc(fields ...string) *Query {
	for _, field := range fields {
		this.Order(field, QueryOrderDesc)
	}
	return this
}

// 按照主键倒排序
func (this *Query) DescPk() *Query {
	this.Order(this.pkName, QueryOrderDesc)
	return this
}

// 设置单个联合查询条件
func (this *Query) Join(dao interface{}, joinType int, on string) *Query {
	this.joins = append(this.joins, QueryJoin{
		DAO:  dao.(DAOWrapper).Object(),
		Type: joinType,
		On:   on,
	})
	return this
}

// 判断是否有联合查询条件
func (this *Query) HasJoins() bool {
	return len(this.joins) > 0
}

// 设置Having条件
func (this *Query) Having(cond string) *Query {
	this.havings = append(this.havings, cond)
	return this
}

// 设置where条件
// @TODO 支持Query、SQL
func (this *Query) Where(wheres ...string) *Query {
	this.wheres = append(this.wheres, wheres...)

	return this
}

// 设置大于条件
func (this *Query) Gt(attr string, value interface{}) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + ">" + ":" + param)

	return this
}

// 设置大于等于条件
func (this *Query) Gte(attr string, value interface{}) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + ">=" + ":" + param)

	return this
}

// 设置小于条件
func (this *Query) Lt(attr string, value interface{}) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "<" + ":" + param)

	return this
}

// 设置小于等于条件
func (this *Query) Lte(attr string, value interface{}) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "<=" + ":" + param)

	return this
}

// 设置不等于条件
func (this *Query) Neq(attr string, value interface{}) *Query {
	var param = "TEA_PARAM_" + this.namedParamPrefix + strconv.Itoa(this.namedParamIndex)
	this.namedParams[param] = value
	this.namedParamIndex++

	this.Where(this.wrapKeyword(attr) + "!=" + ":" + param)

	return this
}

// 设置Group查询条件
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

// 设置like查询条件
// 对表达式自动加上百分号， % ... %
func (this *Query) Like(field string, expr string) *Query {
	wrappedValue, _ := this.wrapAttr("%" + expr + "%")
	this.Where(this.wrapKeyword(field) + " LIKE " + wrappedValue)
	return this
}

// 是否开启SQL Cache
// 只有在SELECT时才有作用
func (this *Query) SQLCache(sqlCache int) *Query {
	this.sqlCache = sqlCache
	return this
}

// 行锁定
func (this *Query) Lock(lock string) *Query {
	this.lock = lock
	return this
}

// 使用索引
func (this *Query) UseIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "USE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// 屏蔽索引
func (this *Query) IgnoreIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "IGNORE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// 强制使用索引
func (this *Query) ForceIndex(index ...string) *Query {
	userIndex := &QueryUseIndex{
		Keyword: "FORCE",
		For:     "",
		Indexes: index,
	}
	this.useIndexes = append(this.useIndexes, userIndex)
	return this
}

// 针对的操作
// 和 UserIndex, IgnoreIndex, ForceIndex 配合使用
func (this *Query) For(clause string) *Query {
	if len(this.useIndexes) == 0 {
		return this
	}

	lastIndex := this.useIndexes[len(this.useIndexes)-1]
	lastIndex.For = clause
	return this
}

// 指定分区
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

// 设置要查询的主键值
func (this *Query) Pk(pks ...interface{}) *Query {
	var realPks = []interface{}{}
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

// 设置between条件
func (this *Query) Between(field string, min interface{}, max interface{}) *Query {
	minValue, _ := this.wrapAttr(min)
	maxValue, _ := this.wrapAttr(max)
	this.Where(field + " BETWEEN " + minValue + " AND " + maxValue)
	return this
}

// 设置一组查询的字段
func (this *Query) Attrs(attrs maps.Map) *Query {
	for key, value := range attrs {
		this.Attr(types.String(key), value)
	}
	return this
}

// 增加某个字段的数值
func (this *Query) Increase(field string, count int) *Query {
	this.savingFields.Put(field, this.wrapKeyword(field)+"+"+this.wrapValue(count))
	return this
}

// 减少某个字段的数值
func (this *Query) Decrease(field string, count int) *Query {
	this.savingFields.Put(field, this.wrapKeyword(field)+"-"+this.wrapValue(count))
	return this
}

// 设置字段值，以便用于删除和修改操作
func (this *Query) Set(field string, value interface{}) *Query {
	this.savingFields.Put(field, this.wrapValue(value))
	return this
}

// 设置一组字段值，以便用于删除和修改操作
// @TODO 需要对keys进行排序
func (this *Query) Sets(values map[string]interface{}) *Query {
	for field, value := range values {
		this.savingFields.Put(field, this.wrapValue(value))
	}
	return this
}

// 设定查询语句中的参数值
// 只有指定where和sql后，才能使用该方法
func (this *Query) Param(name string, value interface{}) *Query {
	this.namedParams[name] = value
	return this
}

// 指定SQL语句
func (this *Query) SQL(sql string) *Query {
	this.sql = sql
	return this
}

// 指定筛选程序
func (this *Query) Filter(filterFn func(one maps.Map) bool) *Query {
	this.filterFn = filterFn
	return this
}

// 指定映射程序
func (this *Query) Map(mapFn func(one maps.Map) maps.Map) *Query {
	this.mapFn = mapFn
	return this
}

// 设置返回的Slice指针
// 在FindAll()方法中生效
func (this *Query) Slice(slicePtr interface{}) *Query {
	this.slicePtr = slicePtr
	return this
}

// 将查询转换为SQL语句
func (this *Query) AsSQL() (string, error) {
	// SQL
	var sql = this.sql

	if len(sql) == 0 {
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
					newResults = append(newResults, this.wrapKeyword(result))
				}
				resultString = strings.Join(newResults, ", ")
			}

			sql = "SELECT"
			if this.sqlCache == QuerySqlCacheOn {
				sql += " SQL_CACHE "
			} else if this.sqlCache == QuerySqlCacheOff {
				sql += " SQL_NO_CACHE "
			}

			sql += "\n  " + resultString + "\n FROM "
			sql += this.wrapTable(this.table)

			// use indexes
			if len(this.useIndexes) > 0 {
				for _, useIndex := range this.useIndexes {
					sql += "\n  " + useIndex.Keyword + " INDEX"
					if len(useIndex.For) > 0 {
						sql += " FOR " + useIndex.For
					}
					quotedIndexes := []string{}
					for _, indexName := range useIndex.Indexes {
						quotedIndexes = append(quotedIndexes, this.wrapKeyword(indexName))
					}
					sql += " (" + strings.Join(quotedIndexes, ", ") + ")"
				}
			}

			// joins
			if len(this.joins) > 0 {
				for _, join := range this.joins {
					if join.Type == QueryJoinDefault {
						sql += ", " + this.wrapTable(join.DAO.Table)
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
						sql += "\n LEFT JOIN " + this.wrapTable(join.DAO.Table)
					case QueryJoinRight:
						sql += "\n RIGHT JOIN " + this.wrapTable(join.DAO.Table)
					}
					sql += this.partitionsSQL()
					if len(join.On) > 0 {
						sql += " ON " + join.On
					}
				}
			} else {
				sql += this.partitionsSQL()
			}
		} else if this.action == QueryActionDelete {
			sql = "DELETE FROM " + this.wrapTable(this.table) + "\n " + this.partitionsSQL()
		} else if this.action == QueryActionUpdate {
			sql = "UPDATE " + this.wrapTable(this.table) + "\n " + this.partitionsSQL() + "SET"
			if this.savingFields.Len() > 0 {
				var mapping = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field interface{}, value interface{}) {
					mapping = append(mapping, this.wrapKeyword(field.(string))+"="+value.(string))
				})
				sql += " " + strings.Join(mapping, ", ")
			}

		} else if this.action == QueryActionInsert {
			sql = "INSERT INTO " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field interface{}, value interface{}) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sql += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}
		} else if this.action == QueryActionReplace {
			sql = "REPLACE " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field interface{}, value interface{}) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sql += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}
		} else if this.action == QueryActionInsertOrUpdate {
			sql = "INSERT INTO " + this.wrapTable(this.table) + "\n" + this.partitionsSQL()
			if this.savingFields.Len() > 0 {
				var fieldNames = []string{}
				var fieldValues = []string{}
				this.savingFields.SortKeys()
				this.savingFields.Range(func(field interface{}, value interface{}) {
					fieldNames = append(fieldNames, this.wrapKeyword(field.(string)))
					fieldValues = append(fieldValues, value.(string))
				})
				sql += " (" + strings.Join(fieldNames, ", ") + ") VALUES (" + strings.Join(fieldValues, ", ") + ")"
			}

			sql += "\nON DUPLICATE KEY UPDATE\n"

			if this.replacingFields.Len() > 0 {
				var mapping = []string{}
				this.replacingFields.SortKeys()
				this.replacingFields.Range(func(field interface{}, value interface{}) {
					mapping = append(mapping, this.wrapKeyword(field.(string))+"="+value.(string))
				})
				sql += strings.Join(mapping, ", ")
			}
		}
	} else {
		sql = this.sql
	}

	// attrs
	var wheres = []string{}
	if this.attrs.Len() > 0 {
		this.attrs.Range(func(_ interface{}, placeholder interface{}) {
			wheres = append(wheres, placeholder.(string))
		})
	}

	// where
	if len(this.wheres) > 0 {
		wheres = append(wheres, this.wheres...)
	}
	if this.action != QueryActionInsert && this.action != QueryActionReplace && this.action != QueryActionInsertOrUpdate && len(wheres) > 0 {
		sql += "\n WHERE " + strings.Join(wheres, " AND ")
	}

	// group
	if this.action == QueryActionFind && len(this.groups) > 0 {
		var groupStrings = []string{}
		for _, group := range this.groups {
			groupStrings = append(groupStrings, this.wrapKeyword(group.Field)+" "+this.orderCode(group.Order))
		}
		sql += "\n GROUP BY " + strings.Join(groupStrings, " AND ")
	}

	// having
	if len(this.havings) > 0 {
		sql += "\n HAVING " + strings.Join(this.havings, " AND ")
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
		sql += "\n ORDER BY " + strings.Join(orderStrings, ", ")
	}

	// limit & offset
	if this.subAction == 0 {
		if this.limit > -1 {
			if this.offset > -1 {
				offsetValue, _ := this.wrapAttr(this.offset)
				limitValue, _ := this.wrapAttr(this.limit)
				sql += fmt.Sprintf("\n LIMIT %s,%s", offsetValue, limitValue)
			} else {
				limitValue, _ := this.wrapAttr(this.limit)
				sql += fmt.Sprintf("\n LIMIT %s", limitValue)
			}
		}
	}

	// JOIN
	if len(this.joins) > 0 {
		reg, _ := stringutil.RegexpCompile("\\b" + "self" + "\\s*\\.")
		sql = reg.ReplaceAllString(sql, this.wrapTable(this.table)+".")

		reg, _ = stringutil.RegexpCompile("\\b" + this.model.Type.Name() + "\\s*\\.")
		sql = reg.ReplaceAllString(sql, this.wrapTable(this.table)+".")

		for _, join := range this.joins {
			modelName := join.DAO.modelWrapper.Type.Name()
			reg, _ := stringutil.RegexpCompile("\\b" + modelName + "\\s*\\.")
			sql = reg.ReplaceAllString(sql, this.wrapTable(join.DAO.Table)+".")
		}
	}

	// lock
	if this.action == QueryActionFind && len(this.lock) > 0 {
		sql += "\n " + this.lock
	}

	// 处理:NamedParam
	var resultSQL = sql
	if !this.isSub {
		reg, err := stringutil.RegexpCompile(":(\\w+)")
		if err != nil {
			return "", err
		}
		var matchedParams = reg.FindAllStringSubmatch(sql, -1)
		this.params = []interface{}{}
		if matchedParams != nil && len(matchedParams) > 0 {
			for _, param := range matchedParams {
				value, ok := this.namedParams[param[1]]
				if ok {
					this.params = append(this.params, value)
				} else {
					this.params = append(this.params, nil)
				}
			}
			resultSQL = reg.ReplaceAllString(resultSQL, "?")
		}
	}

	// debug
	if this.debug {
		logs.Debugf("SQL:" + sql)
		logs.Debugf("params:%#v", this.namedParams)
	}

	return resultSQL, nil
}

// 查找一组数据，返回map数据
func (this *Query) FindOnes() (results []maps.Map, columnNames []string, err error) {
	this.action = QueryActionFind
	sql, err := this.AsSQL()
	if err != nil {
		return nil, nil, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return nil, nil, err
	}
	ones, columnNames, err := stmt.FindOnes(this.params...)

	// 执行 filterFn 和 mapFn
	results = []maps.Map{}
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

	return
}

// 查找一行数据
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

// 查询一组数据， 并返回模型数据
func (this *Query) FindAll() ([]interface{}, error) {
	var ones, _, err = this.FindOnes()
	if err != nil {
		return nil, err
	}

	var results = []interface{}{}
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

// 查询单条数据，返回模型对象
func (this *Query) Find() (interface{}, error) {
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

// 查询单个字段值
func (this *Query) FindCol(defaultValue interface{}) (interface{}, error) {
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

// 判断记录是否存在
func (this *Query) Exist() (bool, error) {
	var one, _, err = this.ResultPk().FindOne()
	if err != nil {
		return false, err
	}
	return len(one) > 0, nil
}

// 执行COUNT查询
func (this *Query) Count() (int64, error) {
	return this.CountAttr("*")
}

// 执行Count查询并返回int
func (this *Query) CountInt() (int, error) {
	count, err := this.Count()
	return int(count), err
}

// 对某个字段进行COUNT查询
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

// 执行SUM查询
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

// 执行MIN查询
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

// 执行MAX查询
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

// 执行AVG查询
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

// 执行查询
func (this *Query) Exec() (*Result, error) {
	this.action = QueryActionExec

	sql, err := this.AsSQL()
	if err != nil {
		return nil, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return nil, err
	}

	result, err := stmt.Exec(this.params...)
	if err != nil {
		return nil, err
	}
	return NewResult(result), nil
}

// 执行REPLACE
func (this *Query) Replace() (rowsAffected int64, lastInsertId int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, 0, errors.New("[Query.Replace()]Replacing fields should be set")
	}

	this.action = QueryActionReplace
	sql, err := this.AsSQL()
	if err != nil {
		return 0, 0, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return 0, 0, err
	}

	result, err := stmt.Exec(this.params...)
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

// 执行INSERT
func (this *Query) Insert() (lastInsertId int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, errors.New("[Query.Insert()]inserting fields should be set")
	}

	this.action = QueryActionInsert
	sql, err := this.AsSQL()
	if err != nil {
		return 0, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(this.params...)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// 执行UPDATE
func (this *Query) Update() (rowsAffected int64, err error) {
	if this.savingFields.Len() == 0 {
		return 0, errors.New("[Query.Update()]updating fields should be set")
	}

	this.action = QueryActionUpdate
	sql, err := this.AsSQL()
	if err != nil {
		return 0, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}

	result, err := stmt.Exec(this.params...)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// 插入或更改
// 依据要插入的数据中的unique键来决定是插入数据还是替换数据
func (this *Query) InsertOrUpdate(insertingValues maps.Map, updatingValues maps.Map) (rowsAffected int64, lastInsertId int64, err error) {
	if insertingValues == nil || len(insertingValues) == 0 {
		return 0, 0, errors.New("[Query.InsertOrUpdate()]inserting values should be set")
	}
	if updatingValues == nil || len(updatingValues) == 0 {
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
	sql, err := this.AsSQL()
	if err != nil {
		return 0, 0, err
	}

	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return 0, 0, err
	}

	result, err := stmt.Exec(this.params...)
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

// 执行DELETE
func (this *Query) Delete() (rowsAffected int64, err error) {
	this.action = QueryActionDelete
	sql, err := this.AsSQL()
	if err != nil {
		return 0, err
	}
	var stmt *Stmt
	if this.canReuse {
		stmt, err = this.db.PrepareOnce(sql)
	} else {
		stmt, err = this.db.Prepare(sql)
		defer stmt.Close()
	}
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(this.params...)
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return rows, nil
}

func (this *Query) stringValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}

	var valueType = reflect.ValueOf(value)
	if valueType.Kind() == reflect.Bool {
		if value.(bool) {
			return 1
		}
		return 0
	}

	if valueType.Kind() == reflect.Slice {
		return string(value.([]byte))
	}

	return fmt.Sprintf("%v", value)
}

// 包装值
func (this *Query) wrapAttr(value interface{}) (placeholder string, isArray bool) {
	switch value := value.(type) {
	case SQL:
		return string(value), false
	case *DBFunc:
		return value.prepareForQuery(this), false
	case *Query:
		value.isSub = true
		sql, err := value.AsSQL()
		if err != nil {
			logs.Errorf("%s", err.Error())
			return
		}

		for paramName, paramValue := range value.namedParams {
			this.namedParams[paramName] = paramValue
			this.namedParamIndex++
		}

		return "IN (" + sql + ")", true
	case *lists.List:
		return this.wrapAttr(value.Slice)
	}

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
func (this *Query) wrapValue(value interface{}) (placeholder string) {
	var valueType = reflect.TypeOf(value)
	var valueTypeName = valueType.Name()

	if valueTypeName == "SQL" {
		return string(value.(SQL))
	}

	if valueType.Kind() == reflect.Slice {
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
	var reg, err = stringutil.RegexpCompile("^\\w+$")
	if err != nil {
		logs.Fatalf("%s", err.Error())
	}
	if !reg.MatchString(keyword) {
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
	var reg, err = stringutil.RegexpCompile("^\\w+$")
	if err != nil {
		logs.Fatalf("%s", err.Error())
		return ""
	}
	if !reg.MatchString(keyword) {
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
func (this *Query) copyModelValue(valueType reflect.Type, data maps.Map) interface{} {
	var pointerValue = reflect.New(valueType)
	var value = reflect.Indirect(pointerValue)
	for index, fieldName := range this.model.Fields {
		var fieldValue, ok = data[fieldName]
		if !ok {
			continue
		}
		if fieldValue == nil {
			continue
		}
		value.Field(index).Set(reflect.ValueOf(this.model.convertValue(fieldValue, this.model.Kinds[index])))
	}
	return pointerValue.Interface()
}
