package dbs

import (
	"database/sql"
	"errors"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type DB struct {
	id     string
	config *DBConfig
	sqlDB  *sql.DB

	stmtManager *StmtManager
}

var dbInitOnce = sync.Once{}
var dbConfig = &Config{}
var dbCachedFactory = map[string]*DB{} // ID => DB Instance
var dbCacheMutex = &sync.Mutex{}

func GlobalConfig() *Config {
	return dbConfig
}

// Default 默认的数据库实例
func Default() (*DB, error) {
	loadConfig()

	defaultDB := dbConfig.Default.DB
	if len(defaultDB) == 0 {
		// 默认为当前的系统环境
		defaultDB = Tea.Env
	}
	return Instance(defaultDB)
}

// Instance 根据ID获取数据库实例
// 如果上下文中已经获得了一个实例，则返回此实例
func Instance(dbId string) (*DB, error) {
	dbCacheMutex.Lock()
	defer dbCacheMutex.Unlock()

	cachedDb, ok := dbCachedFactory[dbId]
	if ok {
		return cachedDb, nil
	}

	loadConfig()

	var db = &DB{
		stmtManager: NewStmtManager(),
	}
	db.id = dbId
	err := db.init()

	if err == nil {
		dbCachedFactory[dbId] = db
	}

	return db, err
}

// NewInstance 根据ID获取一个新的数据库实例
// 不会从上下文的缓存中读取
func NewInstance(dbId string) (*DB, error) {
	loadConfig()

	var db = &DB{
		stmtManager: NewStmtManager(),
	}
	db.id = dbId
	err := db.init()

	return db, err
}

// NewInstanceFromConfig 从配置中构造实例
func NewInstanceFromConfig(config *DBConfig) (*DB, error) {
	sqlDb, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		return nil, err
	}

	// 配置
	if config.Connections.Pool > 0 {
		sqlDb.SetMaxIdleConns(config.Connections.Pool)
	} else {
		sqlDb.SetMaxIdleConns(64)
	}
	if config.Connections.Max > 0 {
		sqlDb.SetMaxOpenConns(config.Connections.Max)
	} else {
		sqlDb.SetMaxOpenConns(128)
	}
	if config.Connections.LifeDuration > 0 {
		sqlDb.SetConnMaxLifetime(config.Connections.LifeDuration)
	}
	sqlDb.SetConnMaxIdleTime(2 * time.Minute)

	var db = &DB{
		stmtManager: NewStmtManager(),
	}

	// close when finalize
	runtime.SetFinalizer(db, func(db *DB) {
		_ = db.Close()
	})

	db.config = config
	db.sqlDB = sqlDb

	// setup stmt manager
	var maxStmtCount = db.queryMaxPreparedStmtCount()
	if maxStmtCount > 0 {
		db.stmtManager.SetMaxCount(maxStmtCount / 16)
	}

	return db, nil
}

func loadConfig() {
	dbInitOnce.Do(func() {
		var dbConfigFile = Tea.ConfigFile("db.yaml")
		var fileBytes, err = os.ReadFile(dbConfigFile)
		if err != nil {
			dbConfigFile = Tea.ConfigFile("db.conf")
			fileBytes, err = os.ReadFile(dbConfigFile)
			if err != nil {
				logs.Errorf("[DB]%s", err.Error())
				return
			}
		}
		err = yaml.Unmarshal(fileBytes, &dbConfig)
		if err != nil {
			logs.Errorf("[DB]%s", err.Error())
			return
		}

		if dbConfig.DBs != nil {
			for key, config := range dbConfig.DBs {
				if len(config.Connections.Life) > 0 {
					duration, err := time.ParseDuration(config.Connections.Life)
					if err != nil {
						logs.Errorf("[DB]wrong connection life:%s", config.Connections.Life)
					} else {
						config.Connections.LifeDuration = duration
						dbConfig.DBs[key] = config
					}
				} /** else {
					// 默认值
					config.Connections.LifeDuration = 30 * time.Second
				}**/
			}
		}
	})
}

func (this *DB) init() error {
	if dbConfig.DBs == nil {
		return errors.New("[DB]fail to load db configuration")
	}

	// 取得配置
	config, ok := dbConfig.DBs[this.id]
	if !ok {
		return errors.New("can not find configuration for '" + this.id + "'")
	}
	this.config = config

	sqlDb, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		return err
	}

	// 配置
	if config.Connections.Pool > 0 {
		sqlDb.SetMaxIdleConns(config.Connections.Pool)
	} else {
		sqlDb.SetMaxIdleConns(64)
	}
	if config.Connections.Max > 0 {
		sqlDb.SetMaxOpenConns(config.Connections.Max)
	} else {
		sqlDb.SetMaxOpenConns(128)
	}
	if config.Connections.LifeDuration > 0 {
		sqlDb.SetConnMaxLifetime(config.Connections.LifeDuration)
	}
	sqlDb.SetConnMaxIdleTime(2 * time.Minute)

	this.sqlDB = sqlDb

	// setup stmt manager
	if this.stmtManager == nil {
		this.stmtManager = NewStmtManager()
	}

	var maxStmtCount = this.queryMaxPreparedStmtCount()
	if maxStmtCount > 0 {
		this.stmtManager.SetMaxCount(maxStmtCount / 16)
	}

	return nil
}

func (this *DB) Config() (*DBConfig, error) {
	if this.config != nil {
		return this.config, nil
	}

	// 根据ID取得配置
	config, ok := dbConfig.DBs[this.id]
	if !ok {
		return config, errors.New("can not find configuration for '" + this.id + "'")
	}

	return config, nil
}

func (this *DB) SetConfig(config *DBConfig) {
	dbConfig.DBs[this.id] = config
}

func (this *DB) Driver() string {
	if this.config != nil {
		return this.config.Driver
	}
	config, ok := dbConfig.DBs[this.id]
	if ok {
		return config.Driver
	}
	return ""
}

func (this *DB) Id() string {
	return this.id
}

func (this *DB) Name() string {
	config, err := this.Config()
	if err != nil {
		return ""
	}
	base := filepath.Base(config.Dsn)
	if len(base) == 0 {
		return ""
	}

	index := strings.Index(base, "?")
	if index < 0 {
		return base
	}
	return base[:index]
}

// Begin 开始一个事务
func (this *DB) Begin() (*Tx, error) {
	tx, err := this.sqlDB.Begin()
	if err != nil {
		return nil, err
	}
	return NewTx(this, tx), nil
}

// StmtManager Get StmtManager
func (this *DB) StmtManager() *StmtManager {
	return this.stmtManager
}

// RunTx 在函数中执行一个事务
func (this *DB) RunTx(callback func(tx *Tx) error) error {
	tx, err := this.Begin()
	if err != nil {
		return err
	}
	if callback != nil {
		err = callback(tx)
	}

	if err != nil {
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}

func (this *DB) Close() error {
	// 关闭语句
	err1 := this.stmtManager.Close()

	// 关闭连接
	err := this.sqlDB.Close()

	return anyError(err, err1)
}

func (db *DB) Exec(query string, params ...interface{}) (sql.Result, error) {
	return db.sqlDB.Exec(query, params...)
}

func (this *DB) Prepare(query string) (*Stmt, error) {
	return this.stmtManager.Prepare(this.sqlDB, query)
}

func (this *DB) PrepareOnce(query string) (*Stmt, bool, error) {
	return this.stmtManager.PrepareOnce(this.sqlDB, query, 0)
}

func (this *DB) FindOnes(query string, args ...interface{}) (results []maps.Map, columnNames []string, err error) {
	stmt, err := this.Prepare(query)
	if err != nil {
		return nil, nil, err
	}

	defer func() {
		_ = stmt.Close()
	}()

	return stmt.FindOnes(args...)
}

func (this *DB) FindPreparedOnes(query string, args ...interface{}) (results []maps.Map, columnNames []string, err error) {
	stmt, cached, err := this.PrepareOnce(query)
	if err != nil {
		return nil, nil, err
	}

	if !cached {
		defer func() {
			_ = stmt.Close()
		}()
	}

	return stmt.FindOnes(args...)
}

func (this *DB) FindOne(query string, args ...interface{}) (maps.Map, error) {
	results, _, err := this.FindOnes(query, args...)
	if err != nil {
		return nil, err
	}

	if len(results) > 0 {
		return results[0], nil
	}
	return nil, nil
}

func (this *DB) FindCol(colIndex int, query string, args ...interface{}) (interface{}, error) {
	stmt, err := this.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = stmt.Close()
	}()

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = rows.Close()
	}()

	columnNames, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var countColumns = len(columnNames)
	if colIndex < 0 || countColumns <= colIndex {
		return nil, nil
	}

	var valuePointers = []interface{}{}
	for i := 0; i < countColumns; i++ {
		var v interface{}
		valuePointers = append(valuePointers, &v)
	}

	if rows.Next() {
		err := rows.Scan(valuePointers...)
		if err != nil {
			return nil, err
		}

		var pointer = valuePointers[colIndex]
		var value = *pointer.(*interface{})

		if value != nil {
			bytes, ok := value.([]byte)
			if ok {
				value = string(bytes)
			}
		}
		return value, nil
	}

	return nil, nil
}

// TableNames 取得所有表格名
func (this *DB) TableNames() ([]string, error) {
	ones, columnNames, err := this.FindOnes("SHOW TABLES")
	if err != nil {
		return nil, err
	}

	var columnName = columnNames[0]
	var results = []string{}
	for _, one := range ones {
		results = append(results, one[columnName].(string))
	}
	return results, nil
}

// FindTable 获取数据表，并包含基本信息
func (this *DB) FindTable(tableName string) (*Table, error) {
	one, err := this.FindOne("SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE table_schema=? AND table_name=?", this.Name(), tableName)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	var table = &Table{
		Name:   tableName,
		Schema: tableName,
		Fields: []*Field{},
	}

	// 表信息
	schema, ok := one["TABLE_SCHEMA"]
	if ok && schema != nil {
		table.Schema = types.String(schema)
	}

	comment, ok := one["TABLE_COMMENT"]
	if ok && comment != nil {
		table.Comment = types.String(comment)
	}

	collation, ok := one["TABLE_COLLATION"]
	if ok && collation != nil {
		table.Collation = types.String(collation)
	}

	engine, ok := one["ENGINE"]
	if ok && engine != nil {
		table.Engine = types.String(engine)
	}

	// 字段信息
	fieldOnes, _, err := this.FindOnes("SHOW FULL COLUMNS FROM `" + tableName + "`")
	if err != nil {
		return table, err
	}
	for _, fieldInfo := range fieldOnes {
		var field = &Field{}

		fieldName, ok := fieldInfo["Field"]
		if ok && fieldName != nil {
			field.Name = types.String(fieldName)
		}

		// not null
		nullValue, ok := fieldInfo["Null"]
		if ok && nullValue != nil {
			field.IsNotNull = nullValue == "NO"
		}

		// primary key
		key, ok := fieldInfo["Key"]
		if ok && key != nil {
			field.IsPrimaryKey = key == "PRI"
		}

		// auto increment
		extra, ok := fieldInfo["Extra"]
		if ok && extra != nil {
			field.AutoIncrement = strings.Contains(types.String(extra), "auto_increment")
			field.Extra = types.String(extra)
		}

		// default value
		defaultValue, ok := fieldInfo["Default"]
		if ok && defaultValue != nil {
			field.DefaultValueString = types.String(defaultValue)
		}

		// comment
		comment, ok := fieldInfo["Comment"]
		if ok && comment != nil {
			field.Comment = types.String(comment)
		}

		// collation
		collation, ok := fieldInfo["Collation"]
		if ok && collation != nil {
			field.Collation = types.String(collation)
		}

		// type
		fullType, ok := fieldInfo["Type"]
		if ok && fullType != nil {
			field.FullType = types.String(fullType)

			// golang data type
			field.parseDataKind()
		}

		table.Fields = append(table.Fields, field)
	}

	return table, nil
}

// FindFullTable 获取数据表，并包含分区，索引信息
func (this *DB) FindFullTable(tableName string) (*Table, error) {
	one, err := this.FindOne("SELECT * FROM INFORMATION_SCHEMA.TABLES WHERE table_schema=? AND table_name=?", this.Name(), tableName)
	if err != nil {
		return nil, err
	}
	if one == nil {
		return nil, nil
	}

	var table = &Table{
		Name:   tableName,
		Schema: tableName,
		Fields: []*Field{},
	}

	// 表信息
	schema, ok := one["TABLE_SCHEMA"]
	if ok && schema != nil {
		table.Schema = types.String(schema)
	}

	comment, ok := one["TABLE_COMMENT"]
	if ok && comment != nil {
		table.Comment = types.String(comment)
	}

	collation, ok := one["TABLE_COLLATION"]
	if ok && collation != nil {
		table.Collation = types.String(collation)
	}

	engine, ok := one["ENGINE"]
	if ok && engine != nil {
		table.Engine = types.String(engine)
	}

	// 字段信息
	fieldOnes, _, err := this.FindOnes("SHOW FULL COLUMNS FROM `" + tableName + "`")
	if err != nil {
		return table, err
	}
	for _, fieldInfo := range fieldOnes {
		var field = &Field{}

		fieldName, ok := fieldInfo["Field"]
		if ok && fieldName != nil {
			field.Name = types.String(fieldName)
		}

		// not null
		nullValue, ok := fieldInfo["Null"]
		if ok && nullValue != nil {
			field.IsNotNull = nullValue == "NO"
		}

		// primary key
		key, ok := fieldInfo["Key"]
		if ok && key != nil {
			field.IsPrimaryKey = key == "PRI"
		}

		// auto increment
		extra, ok := fieldInfo["Extra"]
		if ok && extra != nil {
			field.AutoIncrement = strings.Contains(types.String(extra), "auto_increment")
			field.Extra = types.String(extra)
		}

		// default value
		defaultValue, ok := fieldInfo["Default"]
		if ok && defaultValue != nil {
			field.DefaultValueString = types.String(defaultValue)
		}

		// comment
		comment, ok := fieldInfo["Comment"]
		if ok && comment != nil {
			field.Comment = types.String(comment)
		}

		// collation
		collation, ok := fieldInfo["Collation"]
		if ok && collation != nil {
			field.Collation = types.String(collation)
		}

		// type
		fullType, ok := fieldInfo["Type"]
		if ok && fullType != nil {
			field.FullType = types.String(fullType)

			// golang data type
			field.parseDataKind()
		}

		table.Fields = append(table.Fields, field)
	}

	// 分区
	table.Partitions = []*TablePartition{}
	partitionOnes, _, err := this.FindOnes("SELECT * FROM INFORMATION_SCHEMA.PARTITIONS WHERE TABLE_SCHEMA=? AND TABLE_NAME=?", this.Name(), tableName)
	if err == nil && len(partitionOnes) > 0 {
		for _, one := range partitionOnes {
			partition := &TablePartition{}

			// Name
			name, found := one["PARTITION_NAME"]
			if !found || name == nil {
				continue
			}
			partition.Name = types.String(name)

			// Method
			method, found := one["PARTITION_METHOD"]
			if found && method != nil {
				partition.Method = types.String(method)
			}

			// Expression
			expression, found := one["PARTITION_EXPRESSION"]
			if found && expression != nil {
				partition.Expression = types.String(expression)
			}

			// Description
			description, found := one["PARTITION_DESCRIPTION"]
			if found && description != nil {
				partition.Description = types.String(description)
			}

			// Ordinal Position
			ordinalPosition, found := one["PARTITION_ORDINAL_POSITION"]
			if found && ordinalPosition != nil {
				partition.OrdinalPosition = types.Int(ordinalPosition)
			}

			// Node Group
			nodeGroup, found := one["NODEGROUP"]
			if found && nodeGroup != nil {
				partition.NodeGroup = types.String(nodeGroup)
			}

			// Rows
			rows, found := one["TABLE_ROWS"]
			if found && rows != nil {
				partition.Rows = types.Int64(rows)
			}

			table.Partitions = append(table.Partitions, partition)
		}
	}

	// 索引
	table.Indexes = []*TableIndex{}
	indexOnes, _, err := this.FindOnes("SHOW INDEX FROM `" + tableName + "`")
	if err == nil && len(indexOnes) > 0 {
		for _, one := range indexOnes {
			index := &TableIndex{}

			// name
			keyName, found := one["Key_name"]
			if !found || keyName == nil {
				continue
			}
			index.Name = types.String(keyName)

			// column name
			columnName, found := one["Column_name"]
			if found && columnName != nil {
				columnNameString := types.String(columnName)
				preIndex := table.FindIndexWithName(index.Name)
				if preIndex != nil {
					preIndex.ColumnNames = append(preIndex.ColumnNames, columnNameString)
					continue
				} else {
					index.ColumnNames = []string{columnNameString}
				}
			}

			// unique
			nonUnique, found := one["Non_unique"]
			if found && nonUnique != nil {
				index.IsUnique = types.Int(nonUnique) == 0
			}

			// type
			typeName, found := one["Index_type"]
			if found && typeName != nil {
				index.Type = types.String(typeName)
			}

			// comment
			indexComment, found := one["Index_comment"]
			if found && indexComment != nil {
				index.Comment = types.String(indexComment)
			}

			table.Indexes = append(table.Indexes, index)
		}
	}

	// DDL
	codeOne, err := this.FindOne("SHOW CREATE TABLE `" + tableName + "`")
	if err == nil && codeOne != nil {
		code, found := codeOne["Create Table"]
		if found {
			table.Code = types.String(code)
		}
	}

	return table, nil
}

// FindFunctions 查找所有函数
func (this *DB) FindFunctions() ([]*Function, error) {
	ones, _, err := this.FindOnes("SHOW FUNCTION STATUS WHERE Db='" + this.Name() + "'")
	if err != nil {
		return []*Function{}, err
	}

	functions := []*Function{}
	for _, one := range ones {
		function := &Function{}

		// name
		name, found := one["Name"]
		if found && name != nil {
			function.Name = types.String(name)
		}

		// definer
		definer, found := one["Definer"]
		if found && definer != nil {
			function.Definer = types.String(definer)
		}

		// security type
		securityType, found := one["Security_type"]
		if found && securityType != nil {
			function.SecurityType = types.String(securityType)
		}

		// comment
		comment, found := one["Comment"]
		if found && comment != nil {
			function.Comment = types.String(function)
		}

		// creation
		creationOne, err := this.FindOne("SHOW CREATE FUNCTION `" + function.Name + "`")
		if err == nil {
			creation, found := creationOne["Create Function"]
			if found && creation != nil {
				function.Code = types.String(creation)
			}
		}

		functions = append(functions, function)
	}
	return functions, nil
}

// TablePrefix 取得表前缀
func (this *DB) TablePrefix() string {
	var config, err = this.Config()
	if err != nil {
		return dbConfig.Default.Prefix
	}
	if len(config.Prefix) > 0 {
		return config.Prefix
	}
	return dbConfig.Default.Prefix
}

// Raw 取得原始的数据库连接句柄
func (this *DB) Raw() *sql.DB {
	return this.sqlDB
}

func (this *DB) queryMaxPreparedStmtCount() int {
	// query global variable
	var row = this.sqlDB.QueryRow("SELECT @@max_prepared_stmt_count")
	if row != nil {
		var count int
		err := row.Scan(&count)
		if err == nil && count > 0 {
			return count
		}
	}
	return 0
}
