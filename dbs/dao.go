package dbs

import (
	"errors"
	"github.com/iwind/TeaGo/logs"
	"github.com/iwind/TeaGo/types"
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
	Model        any
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

// Init 初始化
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

// Object 取得封装的对象
func (this *DAOObject) Object() *DAOObject {
	return this
}

// Query 构造查询
func (this *DAOObject) Query(tx *Tx) *Query {
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
		Tx(tx).
		Table(this.Table).
		PkName(this.PkName).
		DAO(this)
}

// Find 查找
func (this *DAOObject) Find(tx *Tx, pk any) (modelPtr any, err error) {
	return this.Query(tx).Pk(pk).Find()
}

// Exist 检查是否存在
func (this *DAOObject) Exist(tx *Tx, pk any) (bool bool, err error) {
	return this.Query(tx).Pk(pk).Exist()
}

// Delete 删除
func (this *DAOObject) Delete(tx *Tx, pk any) (rowsAffected int64, err error) {
	return this.Query(tx).Pk(pk).Delete()
}

// Save 保存
func (this *DAOObject) Save(tx *Tx, operatorPtr any) (err error) {
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

	var query = this.Query(tx)
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
		if !hasPk && fieldName == "created_at" && fieldValue.Interface() == nil {
			var unixTime = time.Now().Unix()
			query.Set("created_at", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if !hasPk && fieldName == "createdAt" && fieldValue.Interface() == nil {
			var unixTime = time.Now().Unix()
			query.Set("createdAt", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if hasPk && fieldName == "updated_at" && fieldValue.Interface() == nil {
			var unixTime = time.Now().Unix()
			query.Set("updated_at", unixTime)
			fieldValue.Set(reflect.ValueOf(unixTime).Convert(fieldValue.Type()))
			continue
		}
		if hasPk && fieldName == "updatedAt" && fieldValue.Interface() == nil {
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
			return err
		}
		if len(this.pkAttr) > 0 {
			pkTypeValue.Set(reflect.ValueOf(lastId).Convert(pkTypeValue.Type()))
		}
	}

	return err
}

// SaveInt64 保存并返回整型ID
func (this *DAOObject) SaveInt64(tx *Tx, operatorPtr any) (pkValue int64, err error) {
	err = this.Save(tx, operatorPtr)
	if err != nil {
		return 0, err
	}
	var modelValue = reflect.Indirect(reflect.ValueOf(operatorPtr))
	pkValueObj := modelValue.FieldByName(this.pkAttr).Interface()
	return types.Int64(pkValueObj), nil
}

// OnInsert 添加Insert回调
func (this *DAOObject) OnInsert(callback func() error) {
	this.insertCallbacks = append(this.insertCallbacks, callback)
}

// NotifyInsert 触发Insert回调
func (this *DAOObject) NotifyInsert() error {
	for _, c := range this.insertCallbacks {
		err := c()
		if err != nil {
			return err
		}
	}
	return nil
}

// OnDelete 添加Delete回调
func (this *DAOObject) OnDelete(callback func() error) {
	this.deleteCallbacks = append(this.deleteCallbacks, callback)
}

// NotifyDelete 触发Delete回调
func (this *DAOObject) NotifyDelete() error {
	for _, c := range this.deleteCallbacks {
		err := c()
		if err != nil {
			return err
		}
	}
	return nil
}

// OnUpdate 添加Update回调
func (this *DAOObject) OnUpdate(callback func() error) {
	this.updateCallbacks = append(this.updateCallbacks, callback)
}

// NotifyUpdate 触发Update回调
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

// NewDAO 初始化DAO
func NewDAO(daoPointer any) any {
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
