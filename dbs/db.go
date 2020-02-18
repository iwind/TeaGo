package dbs

import (
	"database/sql"
	"errors"
	"github.com/go-yaml/yaml"
	"github.com/iwind/TeaGo/Tea"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/maps"
	"github.com/iwind/TeaGo/types"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"time"
)

type DB struct {
	id    string
	sqlDB *sql.DB
	tx    *sql.Tx

	dbStatements map[string]*Stmt // query => stmt
	dbStmtMux    *sync.Mutex

	txStatements map[string]*Stmt // query => stmt
	txStmtMux    *sync.Mutex
}

var dbInitOnce = sync.Once{}
var dbConfig = Config{}
var dbCachedFactory = map[string]*DB{} // ID => DB Instance
var dbCacheMutex = &sync.Mutex{}

// 默认的数据库实例
func Default() (*DB, error) {
	loadConfig()

	if len(dbConfig.Default.DB) > 0 {
		return Instance(dbConfig.Default.DB)
	}

	return nil, errors.New("[DB]there is no db configurations")
}

// 根据ID获取数据库实例
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
		dbStmtMux: &sync.Mutex{},
		txStmtMux: &sync.Mutex{},
	}
	db.id = dbId
	err := db.init()

	if err == nil {
		dbCachedFactory[dbId] = db
	}

	return db, err
}

// 根据ID获取一个新的数据库实例
// 不会从上下文的缓存中读取
func NewInstance(dbId string) (*DB, error) {
	loadConfig()

	var db = &DB{
		dbStmtMux: &sync.Mutex{},
		txStmtMux: &sync.Mutex{},
	}
	db.id = dbId
	err := db.init()

	return db, err
}

func loadConfig() {
	dbInitOnce.Do(func() {
		var dbConfigFile = Tea.ConfigFile("db.yaml")
		var fileBytes, err = ioutil.ReadFile(dbConfigFile)
		if err != nil {
			dbConfigFile = Tea.ConfigFile("db.conf")
			fileBytes, err = ioutil.ReadFile(dbConfigFile)
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

	sqlDb, err := sql.Open(config.Driver, config.Dsn)
	if err != nil {
		logs.Errorf("DB.init():%s", err.Error())
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

	this.sqlDB = sqlDb
	this.dbStatements = map[string]*Stmt{}
	this.txStatements = map[string]*Stmt{}
	return nil
}

func (this *DB) Config() (DBConfig, error) {
	// 取得配置
	config, ok := dbConfig.DBs[this.id]
	if !ok {
		return config, errors.New("can not find configuration for '" + this.id + "'")
	}

	return config, nil
}

func (this *DB) SetConfig(config DBConfig) {
	dbConfig.DBs[this.id] = config
}

func (this *DB) Driver() string {
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

func (this *DB) Begin() error {
	tx, err := this.sqlDB.Begin()
	if err != nil {
		logs.Errorf("DB.Begin():%s", err.Error())
		return err
	}
	this.tx = tx
	return nil
}

func (this *DB) Commit() error {
	if this.tx != nil {
		var err = this.tx.Commit()
		this.tx = nil
		this.txStatements = map[string]*Stmt{}
		return err
	}
	return errors.New("should begin transaction at first")
}

func (this *DB) Rollback() error {
	if this.tx != nil {
		var err = this.tx.Rollback()
		this.tx = nil
		this.txStatements = map[string]*Stmt{}
		return err
	}
	return errors.New("should begin transaction at first")
}

func (this *DB) Close() error {
	err := this.sqlDB.Close()

	/**dbCacheMutex.Lock()
	delete(dbCachedFactory, db.id)
	dbCacheMutex.Unlock()**/

	return err
}

func (db *DB) Exec(query string, params ...interface{}) (sql.Result, error) {
	if db.tx == nil {
		return db.sqlDB.Exec(query, params...)
	} else {
		return db.tx.Exec(query, params...)
	}
}

func (this *DB) Prepare(query string) (*Stmt, error) {
	if this.tx == nil {
		return BuildStmt(this.sqlDB.Prepare(query))
	} else {
		return BuildStmt(this.tx.Prepare(query))
	}
}

func (this *DB) PrepareOnce(query string) (*Stmt, error) {
	if this.tx == nil {
		this.dbStmtMux.Lock()
		defer this.dbStmtMux.Unlock()

		var stmt, ok = this.dbStatements[query]
		if ok {
			return stmt, nil
		}

		sqlStmt, err := this.sqlDB.Prepare(query)
		if err != nil {
			logs.Errorf("DB.PrepareOnce():%s, query:%s", err.Error(), query)
			return BuildStmt(sqlStmt, err)
		}

		stmt, _ = BuildStmt(sqlStmt, nil)

		this.dbStatements[query] = stmt

		return stmt, nil
	} else {
		this.txStmtMux.Lock()
		defer this.txStmtMux.Unlock()

		var stmt, ok = this.txStatements[query]
		if ok {
			return stmt, nil
		}

		sqlStmt, err := this.tx.Prepare(query)
		if err != nil {
			logs.Errorf("DB.PrepareOnce():%s, query:%s", err.Error(), query)
			return BuildStmt(sqlStmt, err)
		}

		stmt, _ = BuildStmt(sqlStmt, nil)

		this.txStatements[query] = stmt
		return stmt, nil
	}
}

func (this *DB) FindOnes(query string, args ...interface{}) (results []maps.Map, columnNames []string, err error) {
	stmt, err := this.Prepare(query)
	if err != nil {
		logs.Errorf("DB.FindOnes():%s", err.Error())
		return nil, nil, err
	}

	defer stmt.Close()

	return stmt.FindOnes(args...)
}

func (this *DB) FindOne(query string, args ...interface{}) (maps.Map, error) {
	results, _, err := this.FindOnes(query, args...)
	if err != nil {
		logs.Errorf("DB.FindOne():%s", err.Error())
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
		logs.Errorf("DB.FindCol():%s", err.Error())
		return nil, err
	}

	defer stmt.Close()

	rows, err := stmt.Query(args...)
	if err != nil {
		logs.Errorf("DB.FindCol():%s", err.Error())
		return nil, err
	}

	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		logs.Errorf("DB.FindCol():%s", err.Error())
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
			logs.Errorf("DB.FindCol():%s", err.Error())
			return nil, err
		}

		var pointer = valuePointers[colIndex]
		var value = *pointer.(*interface{})

		if value != nil {
			var valueType = reflect.TypeOf(value).Kind()
			if valueType == reflect.Slice {
				value = string(value.([]byte))
			}
		}
		return value, nil
	}

	return nil, nil
}

// 取得所有表格名
func (this *DB) TableNames() ([]string, error) {
	ones, columnNames, err := this.FindOnes("SHOW TABLES")
	if err != nil {
		logs.Errorf("DB.TableNames():%s", err.Error())
		return nil, err
	}

	var columnName = columnNames[0]
	var results = []string{}
	for _, one := range ones {
		results = append(results, one[columnName].(string))
	}
	return results, nil
}

// 获取数据表，并包含基本信息
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

// 获取数据表，并包含分区，索引信息
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

// 查找所有函数
func (this *DB) FindFunctions() ([]*Function, error) {
	ones, _, err := this.FindOnes("SHOW FUNCTION STATUS")
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

// 取得表前缀
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

// 取得原始的数据库连接句柄
func (this *DB) OriginDB() *sql.DB {
	return this.sqlDB
}
