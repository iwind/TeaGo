package dbs

import (
	"testing"
	"reflect"
	_ "github.com/go-sql-driver/mysql"
	"encoding/json"
	"time"
	"github.com/iwind/TeaGo/utils/time"
	"github.com/iwind/TeaGo/logs"
)

type TestUser struct {
	ID        int    `field:"id"`
	Gender    int    `field:"gender"`
	Name      string `field:"name"`
	CreatedAt int64  `field:"created_at"`
	Birthday  string `field:"birthday"`
	Mobile    string `field:"mobile"`
}

type TestDb struct {
	Id      uint32 `field:"id"`
	UserId  uint32 `field:"userId"`
	Name    string `field:"name"`
	Comment string `field:"comment"`
	State   uint8  `field:"state"`
}

type TestDbDAO DAO

func (user *TestUser) CreatedDate() string {
	return timeutil.Format("Y-m-d H:i:s", time.Unix(user.CreatedAt, 0))
}

func TestQueryString(t *testing.T) {
	var query = Query{}
	t.Log(query.stringValue(1))
	t.Log(query.stringValue(1.2))
	t.Log(query.stringValue("hello"))
	t.Log(query.stringValue(false), query.stringValue(true))
	t.Log(query.stringValue([]byte("Hello")))
	t.Log(query.stringValue(nil))
}

func TestQuerySQL(t *testing.T) {
	var sql = SQL("SELECT id FROM users")
	t.Log("String:", sql, "Type:", reflect.ValueOf(sql).Type().Name())
}

func TestQuerySlice(t *testing.T) {
	var s = []interface{}{"Hello", 123, false, 1.234}
	var v = reflect.ValueOf(s)

	t.Log(v.Kind())
}

func TestQuery_AsSQL(t *testing.T) {
	var db, err = Instance("db2")
	if err != nil {
		t.Fatal(err)
	}

	var query = NewQuery(nil)
	query.Table("pp_users")
	query.DB(db)
	query.action = QueryActionFind

	query.Where("name=:name AND age=:age").
		Where("created_at>0")
	query.Param("name", "liu").
		Param("age", 20)

	query.Attr("state", 1).State(2)

	query.Limit(10)
	//query.Debug(false)

	t.Logf("Attrs:%#v\n", query.attrs)
	t.Logf("NamedParams:%#v\n", query.namedParams)

	sql, err := query.AsSQL()
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("Params:%#v\n", query.params)

	t.Log(sql)
}

func TestQuery_FindOnes(t *testing.T) {
	var now = time.Now()

	var query = setupUserQuery()
	query.Debug(true)

	query.Offset(0)
	query.Limit(10)

	//query.NoPk(false)
	//query.Result("id", "name", "state", "LENGTH(name) AS nameLength")
	query.Result("id", "name", "state", "created_at")
	//query.Result("*")
	//query.ResultPk()

	//query.State(1)
	//query.Attr("state", 1)
	//query.Attr("name", "刘祥超")
	//query.Like("name", "张三")

	//query.SQLCache(QUERY_SQL_CACHE_OFF)
	//query.Lock(QUERY_LOCK_SHARE_MODE)
	//query.Partitions("p1", "p2")
	//query.Between("id", 1, 3)
	/**query.Attrs(map[string]interface{}{
		"id": 1,
		"state": 1,
	})**/
	query.Pk(1, 2, 3, 100)
	//query.Attr("name", []string{ "行行行店铺", "张三", "刘祥超" })
	//query.Attr("name", SQL("'abc'"))

	//query.Order("id", QUERY_ORDER_DESC)
	//query.Desc("id")
	//query.Desc("id")
	//query.Desc("LENGTH(id)")
	query.AscPk()

	//query.Group("state", QUERY_ORDER_DEFAULT)
	//query.Group("state", QUERY_ORDER_DESC)
	//query.Having("state>0")

	query.action = QueryActionFind
	query.Debug(false)
	//t.Log(query.AsSQL())
	//return

	results, _, err := query.FindOnes()
	if err != nil {
		sql, _ := query.AsSQL()
		t.Fatal(err.Error() + "\nSQL:" + sql)
	}
	jsonBytes, err := json.MarshalIndent(results, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonBytes))
	t.Log(query.AsSQL())
	now = time.Now()
	t.Log(query.FindOnes())
	t.Log(float64(time.Since(now).Nanoseconds()) / 1000000)
}

func TestQuery_FindOne(t *testing.T) {
	var query = setupUserQuery()
	query.Debug(false)

	query.Result("id", "name", "gender", "state")
	query.Where("id=100")

	one, columnNames, err := query.FindOne()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(columnNames)
	outputQueryResult(t, one)
}

func TestQuery_FindCol(t *testing.T) {
	var query = setupUserQuery()
	query.Attr("name", "刘祥超")
	//query.ResultPk()
	query.Result("name", "gender")

	value, err := query.FindCol(0)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(value)
}

func TestQuery_Exist(t *testing.T) {
	var query = setupUserQuery()
	query.Where("id>100")
	t.Log(query.Exist())
}

func TestQuery_Count(t *testing.T) {
	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<200")
		t.Log(query.Count())
	}

	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<2000")
		t.Log(query.CountAttr("DISTINCT state"))
	}
}

func TestQuery_Sum(t *testing.T) {
	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<200")
		t.Log(query.Sum("id", 0))
	}

	{
		var query = setupUserQuery()
		query.Where("id=2000000")
		t.Log(query.Sum("id", 0))
	}
}

func TestQuery_Min(t *testing.T) {
	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<200")
		t.Log(query.Min("id", 0))
	}

	{
		var query = setupUserQuery()
		query.Where("id=2000000")
		t.Log(query.Min("id", 10))
	}
}

func TestQuery_Max(t *testing.T) {
	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<200")
		t.Log(query.Max("id", 0))
	}

	{
		var query = setupUserQuery()
		query.Where("id=2000000")
		t.Log(query.Max("id", 10))
	}
}

func TestQuery_Avg(t *testing.T) {
	{
		var query = setupUserQuery()
		query.Where("id>100 AND id<200")
		t.Log(query.Avg("id", 0))
	}

	{
		var query = setupUserQuery()
		query.Where("id=2000000")
		t.Log(query.Avg("id", 10))
	}
}

func TestQuery_FindAll(t *testing.T) {
	var query = setupUserQuery()
	query.Where("id>300 AND id<500")
	query.Limit(5)

	var values, err = query.FindAll()
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range values {
		var user = value.(*TestUser)
		t.Log("User:", user.ID, user.Name, user.Gender, user.CreatedDate())
	}
}

func TestQuery_Find(t *testing.T) {
	var query = setupUserQuery()
	query.Where("id=1024")
	//query.Attr("id", 1024)
	var result, err = query.Find()
	if err != nil {
		t.Fatal(err)
	}

	var user = result.(*TestUser)
	outputQueryResult(t, user)
	outputQueryResult(t, user.CreatedDate())
}

func TestQuery_Exec(t *testing.T) {
	var query = setupUserQuery()
	query.SQL("UPDATE pp_users SET password='1234567' WHERE id=:userId").
		Param("userId", 1024)
	var result, err = query.Exec()
	if err != nil {
		t.Fatal(err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("AffectedRows:", rows)
}

func TestQuery_Delete(t *testing.T) {
	var query = setupUserQuery()
	query.Attr("id", 8270)
	t.Log(query.Delete())
}

func TestQueryInsert(t *testing.T) {
	var query = setupUserQuery()
	query.Set("name", "张三")
	query.Set("birthday", "1999-10-10")
	t.Log(query.Insert())
}

func TestQueryUpdate(t *testing.T) {
	var query = setupUserQuery()
	query.Attr("id", 8271)
	query.Set("mobile", "13800001234")
	t.Log(query.Update())
}

func TestQueryReplace(t *testing.T) {
	var query = setupUserQuery()
	query.Sets(map[string]interface{}{
		"id":  8272,
		"tel": "010-8888888",
	})
	t.Log(query.Replace())
}

func TestQueryInsertOrUpdate(t *testing.T) {
	var query = setupUserQuery()
	var inserting = map[string]interface{}{
		"id":   8273,
		"name": "李白",
	}
	var updating = map[string]interface{}{
		"name":        "李白2",
		"mobile":      "138",
		"version":     SQL("version+1"),
		"count_views": SQL("RAND() * 10000"),
	}
	t.Log(query.InsertOrUpdate(inserting, updating))
}

func TestQueryJoin(t *testing.T) {
	var query = setupUserQuery()
	var dao = NewDAO(&TestDbDAO{
		DAOObject: DAOObject{
			DB:     "db1",
			Table:  "pp_dbs",
			Model:  new(TestDb),
			PkName: "id",
		},
	}).(*TestDbDAO)

	now := time.Now()

	query.Result("self.id, self.createdAt,TestDb.name AS dbName")
	query.Join(dao, QueryJoinRight, "TestUser.id=TestDb.userId")
	query.Where("self.id>0")
	/**query.Filter(func (one map[string]interface{}) bool {
		id, ok := one["id"]
		if !ok {
			return false
		}
		return types.Int32(id) > 0
	})
	query.Map(func (one map[string]interface{}) map[string]interface{} {
		one["id"] = rand.Int()
		return one
	})**/
	//query.Attr("id", setupUserQuery().ResultPk().AsSQL())
	//query.Attr("id", []int{1, 2, 3})
	/**query.Attr("id", setupUserQuery().ResultPk().Where("id>0"))**/
	query.Attr("createdAt", FuncAbs(-1504256734))
	t.Log(query.AsSQL())
	t.Log(query.params)
	ones, _, err := query.FindOnes()
	if err != nil {
		t.Fatal(err)
	}
	_bytes, _ := json.MarshalIndent(ones, "", "   ")
	t.Log(string(_bytes))
	t.Log(float64(time.Since(now).Nanoseconds()) / 1000 / 1000)
}

func TestQuery_UseIndex(t *testing.T) {
	query := setupUserQuery()
	query.UseIndex("a", "b")
	query.UseIndex("c").For(QueryForOrderBy)
	query.UseIndex()
	query.IgnoreIndex("d", "e")
	query.ForceIndex("f")
	t.Log(query.AsSQL())
}

func TestQuery_UseIndex2(t *testing.T) {
	query := setupUserQuery()
	query.UseIndex("id")
	t.Log(query.AsSQL())
}

func TestFuncAbs(t *testing.T) {
	var query = setupUserQuery()
	t.Log(query.Result(FuncAbs(SQL("id"))).DescPk().FindCol(0))
}

func TestFuncRand(t *testing.T) {
	var query = setupUserQuery()
	t.Log(query.Result(FuncRand()).FindCol(0))
}

func TestFuncFindInSet(t *testing.T) {
	var query = setupUserQuery()
	t.Log(query.Result(FuncFindInSet(SQL("id"), "1,2,3")).Pk(4).FindCol(0))
}

func TestFuncFromUnixtime(t *testing.T) {
	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncFromUnixtime(SQL("createdAt"))).Pk(4).FindCol(0))
	}

	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncFromUnixtimeFormat(SQL("createdAt"), "%Y-%m-%d")).Pk(4).FindCol(0))
	}
}

func TestFuncConcat(t *testing.T) {
	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncConcat(SQL("id"), ":", "1", ",", 2, ",", 3.5)).Pk(4).FindCol(0))
	}

	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncConcat(1)).Pk(4).FindCol(0))
	}

	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncConcatWs(", ", SQL("id"), "1", 2, 3.5)).Pk(4).FindCol(0))
	}
}

func TestFuncLpad(t *testing.T) {
	{
		var query = setupUserQuery()
		t.Log(query.Result(FuncLpad("a", "10", 0)).Pk(4).FindCol(0))
	}
}

func setupUserQuery() *Query {
	db, err := Instance("dev")
	if err != nil {
		logs.Errorf(err.Error())
	}

	var query = NewQuery(new(TestUser))
	query.Table("pp_users")
	query.DB(db)
	return query
}

func outputQueryResult(t *testing.T, result interface{}) {
	jsonBytes, err := json.MarshalIndent(result, "", "    ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(jsonBytes))
}
