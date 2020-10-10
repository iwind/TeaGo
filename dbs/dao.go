package dbs

import (
	"errors"
	"github.com/iwind/TeaGo/logs"
	"log"
	"reflect"
	"sync"
	"time"
)

type DAOObject struct {
	Instance *DB

	DB           string
	Table        string
	PkName       string
	Model        interface{}
	pkAttr       string
	modelWrapper *Model
	fields       map[string]*Field

	insertCallbacks []func() error
	deleteCallbacks []func() error
	updateCallbacks []func() error
}

type DAO struct {
	DAOObject
}

type DAOWrapper interface {
	Object() *DAOObject
}

// 初始化
func (this *DAOObject) Init() error {
	// 主键field映射为attr
	if len(this.PkName) == 0 {
		this.PkName = "id"
	}

	this.modelWrapper = NewModel(this.Model)

	// 获取默认值
	if this.fields == nil {
		var db *DB
		var err error

		if this.Instance != nil {
			db = this.Instance
		} else {
			db, err = Instance(this.DB)
			if err != nil {
				return err
			}
			this.Instance = db
		}

		this.fields = map[string]*Field{}
		table, err := db.FindTable(this.Table)
		if err != nil {
			return errors.New("fail to fetch table fields '" + this.Table + " from db '" + this.DB + "'")
		}
		if table == nil {
			return errors.New("can not find table '" + this.Table + "' from db '" + this.DB + "'")
		}
		for _, field := range table.Fields {
			kind, found := this.modelWrapper.KindsMap[field.Name]
			if !found {
				continue
			}
			attr, found := this.modelWrapper.findAttrWithField(field.Name)
			if !found {
				continue
			}
			if field.Name == this.PkName {
				this.pkAttr = attr
			}
			field.DefaultValue = this.modelWrapper.convertValue(field.DefaultValue, kind)
			this.fields[attr] = field
		}
	}

	return nil
}

// 取得封装的对象
func (this *DAOObject) Object() *DAOObject {
	return this
}

// 构造查询
func (this *DAOObject) Query() *Query {
	var db *DB
	var err error
	if this.Instance != nil {
		db = this.Instance
	} else {
		db, err = Instance(this.DB)
		if err != nil {
			log.Fatal(err)
		}
	}

	return NewQuery(this.Model).
		DB(db).
		Table(this.Table).
		PkName(this.PkName).
		DAO(this)
}

// 查找
func (this *DAOObject) Find(pk interface{}) (modelPtr interface{}, err error) {
	return this.Query().Pk(pk).Find()
}

// 检查是否存在
func (this *DAOObject) Exist(pk interface{}) (bool bool, err error) {
	return this.Query().Pk(pk).Exist()
}

// 删除
func (this *DAOObject) Delete(pk interface{}) (rowsAffected int64, err error) {
	return this.Query().Pk(pk).Delete()
}

// 保存
func (this *DAOObject) Save(operatorPtr interface{}) (newOperatorPtr interface{}, err error) {
	var modelValue = reflect.Indirect(reflect.ValueOf(operatorPtr))
	var hasPk = false
	var pkTypeValue reflect.Value
	if len(this.pkAttr) > 0 {
		pkTypeValue = modelValue.FieldByName(this.pkAttr)
		var pkValue = pkTypeValue.Interface()
		if pkValue == nil {
			hasPk = false
		} else {
			var pkKind = reflect.ValueOf(pkValue).Kind()
			switch pkKind {
			case reflect.Bool:
				hasPk = false
			case reflect.Int:
				hasPk = pkValue.(int) > 0
			case reflect.Int8:
				hasPk = pkValue.(int8) > 0
			case reflect.Int16:
				hasPk = pkValue.(int16) > 0
			case reflect.Int32:
				hasPk = pkValue.(int32) > 0
			case reflect.Int64:
				hasPk = pkValue.(int64) > 0
			case reflect.Uint:
				hasPk = pkValue.(uint) > 0
			case reflect.Uint8:
				hasPk = pkValue.(uint8) > 0
			case reflect.Uint16:
				hasPk = pkValue.(uint16) > 0
			case reflect.Uint32:
				hasPk = pkValue.(uint32) > 0
			case reflect.Uint64:
				hasPk = pkValue.(uint64) > 0
			case reflect.String:
				hasPk = len(pkValue.(string)) > 0
			case reflect.Float32:
				hasPk = pkValue.(float32) > 0
			case reflect.Float64:
				hasPk = pkValue.(float64) > 0
			}
		}
	}

	var query = this.Query()
	var countFields = modelValue.NumField()
	var modelType = modelValue.Type()
	for i := 0; i < countFields; i++ {
		var fieldValue = modelValue.Field(i)
		if !fieldValue.IsValid() {
			continue
		}
		field, found := this.fields[modelType.Field(i).Name]
		if !found {
			continue
		}
		var fieldName = field.Name

		// 支持created_at & createdAt & updated_at & updatedAt
		if !hasPk && fieldName == "created_at" {
			var unixTime = time.Now().Unix()
			query.Set("created_at", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if !hasPk && fieldName == "createdAt" {
			var unixTime = time.Now().Unix()
			query.Set("createdAt", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if hasPk && fieldName == "updated_at" {
			var unixTime = time.Now().Unix()
			query.Set("updated_at", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if hasPk && fieldName == "updatedAt" {
			var unixTime = time.Now().Unix()
			query.Set("updatedAt", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}

		// 主键不更改
		if hasPk && fieldName == this.PkName {
			continue
		}

		// 为nil的不更改
		if fieldValue.IsNil() {
			continue
		}
		query.Set(fieldName, fieldValue.Interface())
	}
	if hasPk {
		_, err = query.Pk(pkTypeValue.Interface()).Update()
	} else {
		lastId, err := query.Insert()
		if err != nil {
			return operatorPtr, err
		}
		if len(this.pkAttr) > 0 {
			pkTypeValue.Set(reflect.ValueOf(lastId).Convert(pkTypeValue.Type()))
		}
	}

	return operatorPtr, err
}

// 添加Insert回调
func (this *DAOObject) OnInsert(callback func() error) {
	this.insertCallbacks = append(this.insertCallbacks, callback)
}

// 触发Insert回调
func (this *DAOObject) NotifyInsert() error {
	for _, c := range this.insertCallbacks {
		err := c()
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加Delete回调
func (this *DAOObject) OnDelete(callback func() error) {
	this.deleteCallbacks = append(this.deleteCallbacks, callback)
}

// 触发Delete回调
func (this *DAOObject) NotifyDelete() error {
	for _, c := range this.deleteCallbacks {
		err := c()
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加Update回调
func (this *DAOObject) OnUpdate(callback func() error) {
	this.updateCallbacks = append(this.updateCallbacks, callback)
}

// 触发Update回调
func (this *DAOObject) NotifyUpdate() error {
	for _, c := range this.updateCallbacks {
		err := c()
		if err != nil {
			return err
		}
	}
	return nil
}

var daoMapping = sync.Map{}
var daoMappingLocker = &sync.Mutex{}

// 初始化DAO
func NewDAO(daoPointer interface{}) interface{} {
	daoMappingLocker.Lock()
	defer daoMappingLocker.Unlock()

	// 如果已经在缓存里直接返回
	var pointerType = reflect.TypeOf(daoPointer).String()
	cachedDAO, ok := daoMapping.Load(pointerType)
	if ok {
		return cachedDAO
	}

	// 初始化
	var pointerValue = reflect.ValueOf(daoPointer)
	v := pointerValue.MethodByName("Init").Call([]reflect.Value{})
	if len(v) > 0 {
		v0 := v[0].Interface()
		if v0 != nil {
			err, ok := v0.(error)
			if ok {
				logs.Println("[DAO]" + err.Error())
			}
		}
	}

	daoMapping.Store(pointerType, daoPointer)
	return daoPointer
}
